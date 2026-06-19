package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type Form struct {
	PrivacyData string `json:"privacyData"`
	Content     string `json:"content"`
}

func (h *Handler) HandleFormSubmit(targetSlackChannelID string) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header
		token := header.Get("X-Form-Token")
		
		if token != h.FormToken {
			return c.String(http.StatusUnauthorized, "unauthorized")
		}
		
		var form Form
		if err := json.NewDecoder(c.Request().Body).Decode(&form); err != nil {
			return c.String(http.StatusBadRequest, "invalid request body")
		}

		_, _, err := h.SlackAPI.PostMessage(targetSlackChannelID, slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", form.PrivacyData, false, true),
				nil,
				nil,
			),
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", form.Content, false, true),
				nil,
				nil,
			),
			slack.NewActionBlock("button",
				slack.NewButtonBlockElement("test", "Click Me", slack.NewTextBlockObject("plain_text", "traQへ転送", false, false)),
			),
		))
		if err != nil {
			log.Printf("failed posting message: %v", err)
		}
		return c.String(http.StatusOK, "ok")
	}
}
