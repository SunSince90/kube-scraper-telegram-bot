package main

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Handler handles communication with the telegram bot
type Handler interface {
}

// telegramHandler is in charge of handling communication with the telegram bot
type telegramHandler struct {
	client  *tgbotapi.BotAPI
	updChan tgbotapi.UpdatesChannel
	lock    sync.Mutex
}

// NewHandler returns an instance of the handler
func NewHandler(token string, offset, timeout int, debugMode bool) (Handler, error) {
	l := log.WithField("func", "NewHandler")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = debugMode
	l.Info("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout

	updChan, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	h := &telegramHandler{
		client:  bot,
		updChan: updChan,
	}

	return h, nil
}
