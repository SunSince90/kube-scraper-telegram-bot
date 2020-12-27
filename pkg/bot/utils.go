package bot

import tgbotapi "gopkg.in/telegram-bot-api.v4"

func getTelegramChatType(chat *tgbotapi.Chat) string {
	if chat.IsChannel() {
		return "channel"
	}
	if chat.IsGroup() {
		return "group"
	}
	if chat.IsSuperGroup() {
		return "supergroup"
	}
	return "private"
}
