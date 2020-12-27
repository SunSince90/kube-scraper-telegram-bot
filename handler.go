package main

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	telegramChats = "telegramChats"
)

// Handler handles communication with the telegram bot
type Handler interface {
	ListenForUpdates(ctx context.Context, stopChan chan struct{})
	SendMessage(dest int64, msg string) (err error)
}

// telegramHandler is in charge of handling communication with the telegram bot
type telegramHandler struct {
	mainCtx context.Context
	client  *tgbotapi.BotAPI
	updChan tgbotapi.UpdatesChannel
	fs      FS
	timeout time.Duration
	lock    sync.Mutex
}

// NewHandler returns an instance of the handler
func NewHandler(ctx context.Context, token string, offset, timeout int, debugMode bool, fs FS) (Handler, error) {
	l := log.WithField("func", "NewHandler")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = debugMode
	l.WithField("account", bot.Self.UserName).Info("Authorized on account")

	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout

	updChan, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	h := &telegramHandler{
		client:  bot,
		updChan: updChan,
		fs:      fs,
		mainCtx: ctx,
		timeout: time.Duration(5) * time.Second,
	}

	return h, nil
}

func (t *telegramHandler) ListenForUpdates(ctx context.Context, stopChan chan struct{}) {
	var u tgbotapi.Update
	l := log.WithField("func", "telegramHandler.ListenForUpdates")
	l.Info("listening for updates...")
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

	l.WithFields(logrus.Fields{"from": update.Message.From.FirstName, "chatID": update.Message.Chat.ID}).Info("got message from", update.Message.From.UserName)
	switch update.Message.Text {
	case "/start", "/restart":
		t.startUser(update.Message.Chat)
	case "/stop":
		t.stopUser(update.Message.Chat)
	case "/siti", "/shops":
		// TODO: print available shops
	default:
		t.unrecognizedCommand(update.Message.Chat)
	}
}

func (t *telegramHandler) startUser(chat *tgbotapi.Chat) {
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.startUser", "user": chat.ID, "username": chat.UserName})

	// Get the user
	doc, err := t.fs.GetChat(chat.ID)
	if err != nil {
		l.WithError(err).Error("error while getting user")
		return
	}

	if doc != nil {
		l.Info("user already subscribed, exiting...")
		return
	}

	// Add the user
	if err := t.fs.InsertChat(chat); err != nil {
		l.WithError(err).Error("error while setting user, returning...")
		return
	}

	l.Debug("sending welcome message")

	if err := t.SendMessage(chat.ID, messageWelcome); err != nil {
		l.WithError(err).Error("could not send message")
		return
	}

	l.Debug("user welcomed successfully")
}

func (t *telegramHandler) stopUser(chat *tgbotapi.Chat) {
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.stopUser", "user": chat.ID, "username": chat.UserName})

	// Get the user
	doc, err := t.fs.GetChat(chat.ID)
	if err != nil {
		l.WithError(err).Error("error while getting user")
		return
	}

	if doc == nil {
		l.Info("user wasn't subscribed, exiting...")
		return
	}

	// Removing user
	if err := t.fs.DeleteChat(chat.ID); err != nil {
		l.WithError(err).Error("error while removing user, returning...")
		return
	}

	l.Debug("sending welcome message")

	if err := t.SendMessage(chat.ID, messageStopUser); err != nil {
		l.WithError(err).Error("could not send message")
		return
	}

	l.Debug("user removed successfully")
}

func (t *telegramHandler) unrecognizedCommand(chat *tgbotapi.Chat) {
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.unrecognizedCommand", "user": chat.ID, "username": chat.UserName})
	l.Debug("sending shrug message")

	if err := t.SendMessage(chat.ID, unrecognizedCommandMessage); err != nil {
		l.WithError(err).Error("could not send message")
		return
	}

	l.Debug("user notified successfully")
}

// SendMessage tries to send a message to the specified destination
func (t *telegramHandler) SendMessage(dest int64, msg string) (err error) {
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.SendMessage"})
	destinations := []int64{}
	if dest != 0 {
		destinations = []int64{dest}
	} else {
		ids, err := t.fs.GetAllChatIDs()
		if err != nil {
			return err
		}

		destinations = ids
	}

	errors := 0
	for _, d := range destinations {
		l = l.WithField("dest", d)
		conf := tgbotapi.NewMessage(d, msg)
		_, err = t.client.Send(conf)
		if err != nil {
			l.WithError(err).Error("error while trying to send message")
			errors++
		} else {
			l.Info("message sent")
		}
	}

	if errors == len(destinations) {
		return err
	}

	return nil
}
