package handler

import (
	"context"
	"log"

	"github.com/traPtitech/traq-ws-bot/payload"
)

func (h *Handler) SetupTraqHandlers() {
	h.TraqBot.OnBotMessageStampsUpdated(func(e *payload.BotMessageStampsUpdated) {
		// e.Stamps[i].StampIDにyokunasasouがある場合、traqのメッセージを削除する
		for _, stamp := range e.Stamps {
			if stamp.StampID == "4e7e3747-168a-4249-b485-91fabe390043" {
				_, err := h.TraqBot.API().MessageApi.DeleteMessage(context.Background(), e.MessageID).Execute()
				if err != nil {
					log.Printf("failed deleting message: %v", err)
				}
				break
			}
		}
	})
}
