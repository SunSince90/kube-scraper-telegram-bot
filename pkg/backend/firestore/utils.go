package firestore

import "github.com/SunSince90/telegram-bot-listener/pkg/backend"

func convertToChat(c *chat) *backend.Chat {
	return &backend.Chat{
		ChatID:    c.ChatID,
		Type:      c.Type,
		Username:  c.Username,
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}
}
