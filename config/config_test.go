package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Backup original config.json if it exists (in case it exists in the test directory)
	origExists := true
	origContent, err := os.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			origExists = false
		} else {
			t.Fatalf("failed to read original config.json: %v", err)
		}
	}
	defer func() {
		if origExists {
			_ = os.WriteFile("config.json", origContent, 0644)
		} else {
			_ = os.Remove("config.json")
		}
	}()

	// Test 1: config.json does not exist
	_ = os.Remove("config.json")
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("expected no error when config.json does not exist, got %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config when config.json does not exist, got %v", cfg)
	}

	// Test 2: config.json exists with valid target channel and profiles
	testJSON := `{
		"channels": [
			{"name": "channel-1", "id": "test-channel-id"},
			{"name": "channel-2", "id": "test-channel-id-2", "group_name": "custom-group"}
		],
		"profiles": [
			{"endpoint": "inbox", "slack_channel_id": "C092013511P", "default_channel_name": "channel-1"}
		]
	}`
	err = os.WriteFile("config.json", []byte(testJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write test config.json: %v", err)
	}

	cfg, err = LoadConfig()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Errorf("expected non-nil config, got nil")
	} else {
		if len(cfg.Channels) != 2 {
			t.Errorf("expected 2 channels, got %d", len(cfg.Channels))
		} else {
			if cfg.Channels[0].Name != "channel-1" || cfg.Channels[0].ID != "test-channel-id" || cfg.Channels[0].GroupName != "default" {
				t.Errorf("expected channel-1 with default group_name, got %+v", cfg.Channels[0])
			}
			if cfg.Channels[1].Name != "channel-2" || cfg.Channels[1].ID != "test-channel-id-2" || cfg.Channels[1].GroupName != "custom-group" {
				t.Errorf("expected channel-2 with custom group_name, got %+v", cfg.Channels[1])
			}
		}

		if len(cfg.Profiles) != 1 {
			t.Errorf("expected 1 profile, got %d", len(cfg.Profiles))
		} else {
			p := cfg.Profiles[0]
			if p.Endpoint != "inbox" || p.SlackChannelID != "C092013511P" || p.DefaultChannelName != "channel-1" {
				t.Errorf("unexpected profile content: %+v", p)
			}
		}
	}

	// Test 3: config.json has invalid JSON
	invalidJSON := `{"channels":`
	err = os.WriteFile("config.json", []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config.json: %v", err)
	}

	cfg, err = LoadConfig()
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config for invalid JSON, got %v", cfg)
	}
}

func TestConfigValidate(t *testing.T) {
	// Test case 1: Valid configuration
	validCfg := &Config{
		Channels: []TraqChannel{
			{Name: "channel-1", ID: "id-1"},
		},
		Profiles: []Profile{
			{Endpoint: "inbox", SlackChannelID: "S1", DefaultChannelName: "channel-1"},
		},
	}
	if err := validCfg.Validate(); err != nil {
		t.Errorf("expected no validation error, got %v", err)
	}

	// Test case 2: Invalid configuration (default channel name doesn't exist)
	invalidCfg := &Config{
		Channels: []TraqChannel{
			{Name: "channel-1", ID: "id-1"},
		},
		Profiles: []Profile{
			{Endpoint: "inbox", SlackChannelID: "S1", DefaultChannelName: "non-existent-channel"},
		},
	}
	if err := invalidCfg.Validate(); err == nil {
		t.Errorf("expected validation error, got nil")
	}
}

func TestLoadConfigRemote(t *testing.T) {
	origRemote := os.Getenv("REMOTE_CONFIG_URL")
	defer os.Setenv("REMOTE_CONFIG_URL", origRemote)

	// Test 1: Fetch valid remote config
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"channels": [
				{"name": "channel-remote", "id": "remote-channel-id"}
			],
			"profiles": [
				{"endpoint": "inbox-remote", "slack_channel_id": "C092013511P", "default_channel_name": "channel-remote"}
			]
		}`))
	}))
	defer ts.Close()

	os.Setenv("REMOTE_CONFIG_URL", ts.URL)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected non-nil config, got nil")
	}
	if len(cfg.Channels) != 1 || cfg.Channels[0].Name != "channel-remote" || cfg.Channels[0].GroupName != "default" {
		t.Errorf("unexpected remote config channels: %+v", cfg.Channels)
	}

	// Test 2: Remote server returns error (e.g. 500)
	tsErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsErr.Close()

	os.Setenv("REMOTE_CONFIG_URL", tsErr.URL)
	cfg, err = LoadConfig()
	if err == nil {
		t.Errorf("expected error when server returns 500, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config when server returns error, got %+v", cfg)
	}

	// Test 3: Remote server returns invalid JSON
	tsInvalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"channels":`))
	}))
	defer tsInvalid.Close()

	os.Setenv("REMOTE_CONFIG_URL", tsInvalid.URL)
	cfg, err = LoadConfig()
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}
