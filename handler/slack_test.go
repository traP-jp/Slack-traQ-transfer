package handler

import (
	"testing"

	"github.com/slack-go/slack"
)

func TestExtractTextFromMessage(t *testing.T) {
	h := &Handler{
		SlackUserID: "U12345",
		SlackBotID:  "B67890",
	}

	// Case 1: Message sent by self (bot user ID)
	msgSelfUser := slack.Message{
		Msg: slack.Msg{
			User: "U12345",
			Blocks: slack.Blocks{
				BlockSet: []slack.Block{
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "privacy data", false, false), nil, nil),
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "actual content 1", false, false), nil, nil),
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "actual content 2", false, false), nil, nil),
				},
			},
		},
	}
	text := h.extractTextFromMessage(msgSelfUser)
	expected := "actual content 1\nactual content 2"
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}

	// Case 2: Message sent by self (bot ID)
	msgSelfBot := slack.Message{
		Msg: slack.Msg{
			BotID: "B67890",
			Blocks: slack.Blocks{
				BlockSet: []slack.Block{
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "privacy data", false, false), nil, nil),
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "actual content 1", false, false), nil, nil),
				},
			},
		},
	}
	text = h.extractTextFromMessage(msgSelfBot)
	expected = "actual content 1"
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}

	// Case 3: Message sent by someone else (another user)
	msgOtherUser := slack.Message{
		Msg: slack.Msg{
			User: "U99999",
			Blocks: slack.Blocks{
				BlockSet: []slack.Block{
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "first block content", false, false), nil, nil),
					slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "second block content", false, false), nil, nil),
				},
			},
		},
	}
	text = h.extractTextFromMessage(msgOtherUser)
	expected = "first block content\nsecond block content"
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}
