package main

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Handler handles communication with the telegram bot
type Handler interface {
	ListenForUpdates(context.Context, chan struct{})
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

func (t *telegramHandler) ListenForUpdates(ctx context.Context, stopChan chan struct{}) {
	var u tgbotapi.Update
	l := log.WithField("func", "telegramHandler.ListenForUpdates")
	for {
		select {
		case u = <-t.updChan:
			t.parseUpdate(u)
		case <-ctx.Done():
			l.Info("exiting")
			close(stopChan)
			return
		}
	}
}

func (t *telegramHandler) parseUpdate(update tgbotapi.Update) {
	l := log.WithField("func", "telegramHandler.parseUpdate")
	l.Info("got update")

	if update.Message == nil { // ignore any non-Message Updates
		l.Info("got non-message update. Skipping...")
		return
	}

	l.Info("got message from", update.Message.From.UserName)
	switch update.Message.Text {
	case "/start", "/restart":
		// TODO: t.addNewUser()
	case "/stop":
		// TODO: t.removeUser()
	case "/siti", "/shops":
		// TODO: print available shops
	default:
		// TODO: command not recognized
	}
}
