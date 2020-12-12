package main

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
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
	fs      *firestore.Client
	timeout time.Duration
	lock    sync.Mutex
}

// NewHandler returns an instance of the handler
func NewHandler(ctx context.Context, token string, offset, timeout int, debugMode bool, fs *firestore.Client) (Handler, error) {
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
		fs:      fs,
		mainCtx: ctx,
		timeout: time.Duration(5) * time.Second,
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
		t.unrecognizedCommand(update.Message.Chat)
	}
}

func (t *telegramHandler) addNewUser(chat *tgbotapi.Chat) {
	l := log.WithFields(logrus.Fields{"func": "telegramHandler.addNewUser", "user": chat.ID, "username": chat.UserName})

	// Get the user
	doc := func() *firestore.DocumentSnapshot {
		docPath := path.Join(telegramChats, fmt.Sprintf("%d", chat.ID))
		ctx, canc := context.WithTimeout(t.mainCtx, t.timeout)
		defer canc()

		docSnap, err := t.fs.Doc(docPath).Get(ctx)
		if err != nil {
			l.WithError(err).Error("error while getting user")
			return nil
		}

		return docSnap
	}()

	if doc.Exists() {
		l.Info("user already subscribed, exiting...")
		return
	}

	// Add the user
	success := func(ref *firestore.DocumentRef) bool {
		ctx, canc := context.WithTimeout(t.mainCtx, t.timeout)
		defer canc()

		addChat := TelegramChat{
			ChatID:    chat.ID,
			Username:  chat.UserName,
			FirstName: chat.FirstName,
			LastName:  chat.LastName,
		}

		_, err := ref.Set(ctx, addChat)
		if err != nil {
			l.Info("error while setting")
			return false
		}
		return true
	}(doc.Ref)
	if !success {
		l.Error("could not set data, returning...")
		return
	}

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
	conf := tgbotapi.NewMessage(dest, msg)
	_, err = t.client.Send(conf)
	return
}
