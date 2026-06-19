package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type TraqChannel struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	GroupName string `json:"group_name"`
}

type Profile struct {
	Endpoint           string `json:"endpoint"`
	SlackChannelID     string `json:"slack_channel_id"`
	DefaultChannelName string `json:"default_channel_name"`
}

type Config struct {
	Channels []TraqChannel `json:"channels"`
	Profiles []Profile     `json:"profiles"`
}

func (c *Config) populateDefaults() {
	for i := range c.Channels {
		if c.Channels[i].GroupName == "" {
			c.Channels[i].GroupName = "default"
		}
	}
}

func (c *Config) Validate() error {
	c.populateDefaults()
	channelNames := make(map[string]bool)
	for _, ch := range c.Channels {
		channelNames[ch.Name] = true
	}
	for _, p := range c.Profiles {
		if !channelNames[p.DefaultChannelName] {
			return fmt.Errorf("default channel %q for endpoint %q does not exist in channels list", p.DefaultChannelName, p.Endpoint)
		}
	}
	return nil
}

func LoadConfig() (*Config, error) {
	if remoteURL := os.Getenv("REMOTE_CONFIG_URL"); remoteURL != "" {
		req, err := http.NewRequest(http.MethodGet, remoteURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for remote config: %w", err)
		}

		if token := os.Getenv("REMOTE_CONFIG_TOKEN"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch remote config: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch remote config: status code %d", resp.StatusCode)
		}

		var cfg Config
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("failed to decode remote config: %w", err)
		}
		cfg.populateDefaults()
		return &cfg, nil
	}

	file, err := os.Open("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	cfg.populateDefaults()
	return &cfg, nil
}
