package main

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

// Handler handles communication with the telegram bot
type Handler interface {
	ListenForUpdates(ctx context.Context, stopChan chan struct{})
	SendMessage(dest int64, msg string) (err error)
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
		t.addNewUser(update.Message.Chat)
	case "/stop":
		t.removeUser(update.Message.Chat)
	case "/siti", "/shops":
		// TODO: print available shops
	default:
		// TODO: command not recognized
	}
}

func (t *telegramHandler) addNewUser(chat *tgbotapi.Chat) {
	// TODO: Check if this user was already added previously
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.addNewUser", "user": chat.ID, "username": chat.UserName})
	l.Debug("sending welcome message")

	if err := t.SendMessage(chat.ID, messageWelcome); err != nil {
		l.WithError(err).Error("could not send message")
		return
	}

	l.Debug("user welcomed successfully")
}

func (t *telegramHandler) removeUser(chat *tgbotapi.Chat) {
	// TODO: Check if this user was already added previously
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.removeUser", "user": chat.ID, "username": chat.UserName})
	l.Debug("sending remove message")

	if err := t.SendMessage(chat.ID, messageRemoveUser); err != nil {
		l.WithError(err).Error("could not send message")
		return
	}

	l.Debug("user notified successfully")
}

// SendMessage tries to send a message to the specified destination
func (t *telegramHandler) SendMessage(dest int64, msg string) (err error) {
	conf := tgbotapi.NewMessage(dest, msg)
	_, err = t.client.Send(conf)
	return
}
