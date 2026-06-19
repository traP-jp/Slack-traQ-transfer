package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/traPtitech/go-traq"
)

func (h *Handler) RunSlackSocketLoop(socket *socketmode.Client) {
	for envelope := range socket.Events {
		switch envelope.Type {
		case socketmode.EventTypeInteractive:
			interaction, ok := envelope.Data.(slack.InteractionCallback)
			if !ok {
				continue
			}
			socket.Ack(*envelope.Request)

			switch interaction.Type {
			case slack.InteractionTypeMessageAction, slack.InteractionTypeBlockActions:
				text := h.extractTextFromMessage(interaction.Message)
				if !h.isSelf(interaction.Message) {
					text = fmt.Sprintf("```text\n%v\n```", text)
				}

				defaultChannelID := h.getDefaultChannelID(interaction.Channel.ID)
				slackChannelName := interaction.Channel.Name

				err := h.openDialog(interaction.TriggerID, text, defaultChannelID, slackChannelName)
				if err != nil {
					log.Printf("failed opening dialog: %v", err)
				}

			case slack.InteractionTypeDialogSubmission:
				targetChannelID := interaction.Submission["target_channel_id"]
				if targetChannelID == "" {
					log.Printf("warning: target_channel_id not specified in dialog submission")
					break
				}

				slackChannelName := interaction.Submission["slack_channel_name"]
				content := interaction.Submission["content"]
				_, _, err := h.TraqBot.API().MessageApi.PostMessage(context.Background(), targetChannelID).PostMessageRequest(traq.PostMessageRequest{
					Content: fmt.Sprintf("**渉外 Slack** [`#%s`] **から転送**\n%v", slackChannelName, content),
					Embed:   func(value bool) *bool { return &value }(true),
				}).Execute()

				if err != nil {
					log.Printf("failed posting message: %v", err)
				}
			}
		}
	}
}

func (h *Handler) getDefaultChannelID(slackChannelID string) string {
	for _, profile := range h.Config.Profiles {
		if profile.SlackChannelID == slackChannelID {
			for _, ch := range h.Config.Channels {
				if ch.Name == profile.DefaultChannelName {
					return ch.ID
				}
			}
		}
	}
	if len(h.Config.Channels) > 0 {
		return h.Config.Channels[0].ID
	}
	return ""
}

func (h *Handler) openDialog(triggerID string, text string, defaultChannelID string, slackChannelName string) error {
	var optionGroups []slack.DialogOptionGroup
	groupMap := make(map[string]int)

	for _, ch := range h.Config.Channels {
		groupName := ch.GroupName
		if groupName == "" {
			groupName = "default"
		}

		idx, exists := groupMap[groupName]
		if !exists {
			idx = len(optionGroups)
			optionGroups = append(optionGroups, slack.DialogOptionGroup{
				Label: groupName,
			})
			groupMap[groupName] = idx
		}

		optionGroups[idx].Options = append(optionGroups[idx].Options, slack.DialogSelectOption{
			Label: ch.Name,
			Value: ch.ID,
		})
	}

	return h.SlackAPI.OpenDialog(triggerID, slack.Dialog{
		TriggerID:   triggerID,
		CallbackID:  "dialog",
		Title:       "Dialog",
		SubmitLabel: "Submit",
		Elements: []slack.DialogElement{
			slack.TextInputElement{
				DialogInput: slack.DialogInput{
					Label:       "Content",
					Name:        "content",
					Placeholder: "Content",
					Type:        slack.InputTypeTextArea,
				},
				Value: text,
			},
			slack.TextInputElement{
				DialogInput: slack.DialogInput{
					Label:       "From",
					Name:        "slack_channel_name",
					Placeholder: "Channel Name",
					Type:        slack.InputTypeText,
				},
				Value: slackChannelName,
			},
			slack.DialogInputSelect{
				DialogInput: slack.DialogInput{
					Label:       "To",
					Name:        "target_channel_id",
					Type:        slack.InputTypeSelect,
					Placeholder: "Select Channel",
				},
				OptionGroups: optionGroups,
				Value:        defaultChannelID,
			},
		},
	})
}

func (h *Handler) extractTextFromMessage(message slack.Message) string {
	if !h.isSelf(message) {
		return message.Text
	}

	var blockTexts []string
	for _, b := range message.Msg.Blocks.BlockSet[1:] {
		if b.BlockType() == slack.MBTSection {
			if section, ok := b.(*slack.SectionBlock); ok && section.Text != nil {
				blockTexts = append(blockTexts, section.Text.Text)
			}
		}
	}

	return strings.Join(blockTexts, "\n\n")
}

func (h *Handler) isSelf(message slack.Message) bool {
	return (message.User != "" && message.User == h.SlackUserID) || (message.BotID != "" && message.BotID == h.SlackBotID)
}
