package handler

import (
	"slack-traq-transfer/config"

	"github.com/slack-go/slack"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

type Handler struct {
	SlackAPI    *slack.Client
	TraqBot     *traqwsbot.Bot
	Config      *config.Config
	FormToken   string
	SlackUserID string
	SlackBotID  string
}

func NewHandler(slackAPI *slack.Client, traqBot *traqwsbot.Bot, cfg *config.Config, formToken string) *Handler {
	return &Handler{
		SlackAPI:  slackAPI,
		TraqBot:   traqBot,
		Config:    cfg,
		FormToken: formToken,
	}
}
