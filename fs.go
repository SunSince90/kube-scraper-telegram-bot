package main

import (
	"context"
	"fmt"
	"path"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FS connects with Firestore and performs operations in it
type FS interface {
	Close()
	GetChat(int64) (*TelegramChat, error)
	InsertChat(*tgbotapi.Chat) error
	DeleteChat(int64) error
}

type fsHandler struct {
	client  *firestore.Client
	app     *firebase.App
	mainCtx context.Context
	timeout time.Duration
}

// NewFSHandler returns a fsHandler, which is an implementation for FS
func NewFSHandler(ctx context.Context, projectName, servAcc string) (FS, error) {
	conf := &firebase.Config{ProjectID: projectName}
	app, err := firebase.NewApp(ctx, conf, option.WithServiceAccountFile(servAcc))
	if err != nil {
		return nil, err
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return &fsHandler{
		mainCtx: ctx,
		app:     app,
		client:  fsClient,
		timeout: time.Duration(5) * time.Second,
	}, nil
}

// GetChat gets a chat from firestore
func (f *fsHandler) GetChat(id int64) (*TelegramChat, error) {
	docPath := path.Join(telegramChats, fmt.Sprintf("%d", id))
	ctx, canc := context.WithTimeout(f.mainCtx, f.timeout)
	defer canc()

	doc, err := f.client.Doc(docPath).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}

		return nil, err
	}

	var chat TelegramChat
	if err := doc.DataTo(&chat); err != nil {
		return nil, err
	}

	return &chat, nil
}

// InsertChat inserts a chat into firestore
func (f *fsHandler) InsertChat(chat *tgbotapi.Chat) error {
	if chat == nil {
		return fmt.Errorf("chat is nil")
	}

	docPath := path.Join(telegramChats, fmt.Sprintf("%d", chat.ID))
	ctx, canc := context.WithTimeout(f.mainCtx, f.timeout)
	defer canc()

	addChat := TelegramChat{
		ChatID:    chat.ID,
		Username:  chat.UserName,
		FirstName: chat.FirstName,
		LastName:  chat.LastName,
	}

	_, err := f.client.Doc(docPath).Set(ctx, addChat)
	return err
}

// DeleteChat deletes a chat from firestore
func (f *fsHandler) DeleteChat(id int64) error {
	docPath := path.Join(telegramChats, fmt.Sprintf("%d", id))
	ctx, canc := context.WithTimeout(f.mainCtx, f.timeout)
	defer canc()

	_, err := f.client.Doc(docPath).Delete(ctx)
	return err
}

// Close the firestore client
func (f *fsHandler) Close() {
	f.client.Close()
}
