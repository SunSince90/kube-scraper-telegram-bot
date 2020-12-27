package bot

import "context"

// Bot is the telegram bot
type Bot interface {
	// ListenForUpdates listens for updates coming from telegram
	ListenForUpdates(context.Context, chan struct{})
}
