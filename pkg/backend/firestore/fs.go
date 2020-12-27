package firestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/SunSince90/telegram-bot-listener/pkg/backend"

	fs "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/rs/zerolog"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	log zerolog.Logger
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

type fsBackend struct {
	cache  map[int64]*backend.Chat
	client *fs.Client
	app    *firebase.App
	*Options
	lock sync.Mutex
}

// NewBackend returns a fsHandler, which is an implementation for FS
func NewBackend(ctx context.Context, servAcc string, opts *Options) (backend.Backend, error) {
	// -- Validation
	if len(opts.ChatsCollection) == 0 {
		return nil, fmt.Errorf("no chat collection set")
	}
	if len(opts.ProjectName) == 0 {
		return nil, fmt.Errorf("no project name set")
	}

	conf := &firebase.Config{ProjectID: opts.ProjectName}
	app, err := firebase.NewApp(ctx, conf, option.WithServiceAccountFile(servAcc))
	if err != nil {
		return nil, err
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	fs := &fsBackend{
		app:     app,
		client:  fsClient,
		Options: opts,
	}

	if opts.UseCache {
		fs.cache = map[int64]*backend.Chat{}
	}

	return fs, nil
}

// Close the client
func (f *fsBackend) Close() {
	f.client.Close()
}

// GetChatByID gets a chat from firestore
func (f *fsBackend) GetChatByID(id int64) (*backend.Chat, error) {
	if id == 0 {
		return nil, fmt.Errorf("chat id cannot be 0")
	}

	l := log.With().Str("func", "GetChatByID").Int64("id", id).Logger()
	if f.UseCache {
		// TODO: implement cache
		l.Debug().Msg("pulled from cache")
	}

	docPath := path.Join(f.ChatsCollection, fmt.Sprintf("%d", id))
	timeout := time.Duration(15) * time.Second
	ctx, canc := context.WithTimeout(context.Background(), timeout)
	defer canc()

	doc, err := f.client.Doc(docPath).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, backend.ErrNotFound
		}

		return nil, err
	}

	var _chat chat
	if err := doc.DataTo(&_chat); err != nil {
		return nil, err
	}
	c := convertToChat(&_chat)

	if f.UseCache {
		// TODO: implement cache
	}

	return c, nil
}

// GetChatByUsername gets a chat from firestore by username
func (f *fsBackend) GetChatByUsername(username string) (*backend.Chat, error) {
	if len(username) == 0 {
		return nil, fmt.Errorf("chat username cannot be 0")
	}

	timeout := time.Duration(15) * time.Second
	ctx, canc := context.WithTimeout(context.Background(), timeout)
	defer canc()

	docIter := f.client.Collection(f.ChatsCollection).Where("username", "==", username).Limit(1).Documents(ctx)
	doc, err := docIter.Next()
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return nil, backend.ErrNotFound
		}

		return nil, err
	}

	var _chat chat
	if err := doc.DataTo(&_chat); err != nil {
		return nil, err
	}

	if f.UseCache {
		// TODO: implement cache
	}

	return convertToChat(&_chat), nil
}

// StoreChats inserts a chat into firestore
func (f *fsBackend) StoreChat(c *backend.Chat) error {
	if c == nil {
		return fmt.Errorf("chat cannot be nil")
	}

	docPath := path.Join(f.ChatsCollection, fmt.Sprintf("%d", c.ChatID))
	timeout := time.Duration(15) * time.Second
	ctx, canc := context.WithTimeout(context.Background(), timeout)
	defer canc()

	addChat := chat{
		ChatID:    c.ChatID,
		Type:      c.Type,
		Username:  c.Username,
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}

	_, err := f.client.Doc(docPath).Set(ctx, addChat)
	if err != nil {
		return err
	}

	if f.UseCache {
		// TODO: implement cache
	}

	return nil
}

// DeleteChat deletes a chat from firestore
func (f *fsBackend) DeleteChat(id int64) error {
	if id == 0 {
		return fmt.Errorf("chat id cannot be 0")
	}

	docPath := path.Join(f.ChatsCollection, fmt.Sprintf("%d", id))
	timeout := time.Duration(15) * time.Second
	ctx, canc := context.WithTimeout(context.Background(), timeout)
	defer canc()

	_, err := f.client.Doc(docPath).Delete(ctx)

	if f.UseCache {
		// TODO: implement cache
	}

	return err
}
