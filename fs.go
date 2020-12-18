package main

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FS connects with Firestore and performs operations in it
type FS interface {
	Close()
	GetChat(int64) (*TelegramChat, error)
	GetAllChatIDs() ([]int64, error)
	InsertChat(*tgbotapi.Chat) error
	DeleteChat(int64) error
}

type fsHandler struct {
	cacheChats map[int64]*TelegramChat
	client     *firestore.Client
	app        *firebase.App
	mainCtx    context.Context
	timeout    time.Duration
	lock       sync.Mutex
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
		mainCtx:    ctx,
		app:        app,
		client:     fsClient,
		timeout:    time.Duration(10) * time.Second,
		cacheChats: map[int64]*TelegramChat{},
	}, nil
}

func (f *fsHandler) getChatFromCache(id int64) *TelegramChat {
	f.lock.Lock()
	defer f.lock.Unlock()

	c, exists := f.cacheChats[id]
	if exists && c != nil {
		return c
	}

	return nil
}

func (f *fsHandler) getAllChatsIDsFromCache() []int64 {
	f.lock.Lock()
	defer f.lock.Unlock()

	if len(f.cacheChats) == 0 {
		return []int64{}
	}

	list := make([]int64, len(f.cacheChats))
	i := 0

	for id := range f.cacheChats {
		list[i] = id
		i++
	}

	return list
}

func (f *fsHandler) insertChatIntoCache(chat *TelegramChat) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.cacheChats[chat.ChatID] = chat
}

func (f *fsHandler) deleteChatFromCache(id int64) {
	f.lock.Lock()
	defer f.lock.Unlock()

	delete(f.cacheChats, id)
}

// GetAllChats gets all chat from firestore
func (f *fsHandler) GetAllChatIDs() ([]int64, error) {
	l := log.WithField("func", "getAllChats").Logger

	if chats := f.getAllChatsIDsFromCache(); len(chats) > 0 {
		return chats, nil
	}

	ctx, canc := context.WithTimeout(f.mainCtx, f.timeout)
	defer canc()

	list := []int64{}
	dociter := f.client.Collection(telegramChats).Documents(ctx)
	defer dociter.Stop()

	for {
		doc, err := dociter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return nil, err
		}

		var chat TelegramChat
		if err := doc.DataTo(&chat); err != nil {
			l.WithField("id", doc.Ref.ID).Info("error while trying to get this document, skipping...")
			continue
		}

		f.insertChatIntoCache(&chat)
	}

	return list, nil
}

// GetChat gets a chat from firestore
func (f *fsHandler) GetChat(id int64) (*TelegramChat, error) {
	if chat := f.getChatFromCache(id); chat != nil {
		return chat, nil
	}

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

	f.insertChatIntoCache(&chat)
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
	if err != nil {
		f.insertChatIntoCache(&addChat)
	}
	return err
}

// DeleteChat deletes a chat from firestore
func (f *fsHandler) DeleteChat(id int64) error {
	defer f.deleteChatFromCache(id)
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
