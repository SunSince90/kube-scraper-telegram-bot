package bot

import (
	"context"
	"fmt"
	"os"
	"sync"

	ksb "github.com/SunSince90/kube-scraper-backend/pkg/backend"
	pb "github.com/SunSince90/kube-scraper-backend/pkg/pb"
	"github.com/rs/zerolog"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var (
	log zerolog.Logger
)

const (
	defaultTimeout int = 3600
	defaultOffset  int = 0
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

type telegramBot struct {
	client  *tgbotapi.BotAPI
	updChan tgbotapi.UpdatesChannel
	texts   map[string]string
	backend ksb.Backend
	lock    sync.Mutex
}

// NewBotListener returns a new instance of the bot listener
func NewBotListener(opts *TelegramOptions, backend ksb.Backend) (Bot, error) {
	// -- Validation
	if backend == nil {
		return nil, fmt.Errorf("backend not set")
	}
	if opts == nil {
		return nil, fmt.Errorf("options not provided")
	}
	if len(opts.Token) == 0 {
		return nil, fmt.Errorf("no token provided")
	}

	l := log.With().Str("func", "NewBotListener").Logger()

	// -- Get the client
	bot, err := tgbotapi.NewBotAPI(opts.Token)
	if err != nil {
		return nil, err
	}

	if opts.Debug != nil && *opts.Debug {
		bot.Debug = *opts.Debug
	}

	l.Debug().Str("account", bot.Self.UserName).Msg("authorized")

	// -- Get the values from the options
	var offset, timeout = defaultOffset, defaultTimeout
	if opts.Offset != nil && *opts.Offset > 0 {
		offset = *opts.Offset
	}
	if opts.Timeout != nil && *opts.Timeout > 0 {
		timeout = *opts.Timeout
	}

	// -- Get the updates channel
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout
	updChan, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	b := &telegramBot{
		client:  bot,
		updChan: updChan,
		texts:   opts.Texts,
		backend: backend,
	}

	return b, nil
}

func (b *telegramBot) ListenForUpdates(ctx context.Context, exitChan chan struct{}) {
	l := log.With().Str("func", "ListenForUpdates").Logger()

	var u tgbotapi.Update
	for {
		select {
		case u = <-b.updChan:
			b.parseUpdate(&u)
		case <-ctx.Done():
			l.Info().Msg("exiting")
			close(exitChan)
			return
		}
	}
}

func (b *telegramBot) parseUpdate(update *tgbotapi.Update) {
	l := log.With().Str("func", "ListenForUpdates").Logger()

	if update.Message == nil {
		// ignore any non-Message Updates
		l.Debug().Msg("got non-message update. Skipping...")
		return
	}

	l = l.With().Str("from", update.Message.From.FirstName).Int64("chat-id", update.Message.Chat.ID).Logger()

	switch update.Message.Text {
	case "/start", "/restart":
		b.startChat(update)
	case "/stop":
		b.stopChat(update)
	case "/siti", "/websites":
		// TODO: print available shops
	default:
		// Nothing is printed if the message is not recognized
	}
}

func (b *telegramBot) startChat(update *tgbotapi.Update) {
	l := log.With().Str("func", "startChat").Logger()
	if b.backend == nil {
		l.Warn().Msg("no backend is set")
		return
	}

	// -- Get the chat
	_, err := b.backend.GetChatByID(update.Message.Chat.ID)
	if err != nil {
		if err != ksb.ErrNotFound {
			l.Err(err).Msg("error while getting chat")
			return
		}

		l.Debug().Msg("chat already exists, returning...")
		return
	}

	// -- Store the chat on firestore
	c := &pb.Chat{
		Id:        update.Message.Chat.ID,
		Title:     update.Message.Chat.Title,
		Type:      getTelegramChatType(update.Message.Chat),
		Username:  update.Message.Chat.UserName,
		FirstName: update.Message.Chat.FirstName,
		// LastName: update.Message.Chat.LastName,
	}
	if err := b.backend.StoreChat(c); err != nil {
		l.Err(err).Msg("error while storing chat on firestore")
		return
	}

	// -- Notify the user
	message := b.texts["messageWelcome"]
	if !update.Message.Chat.IsPrivate() {
		message = b.texts["messageWelcomeGroup"]
	}

	conf := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	conf.ReplyToMessageID = update.Message.MessageID

	if _, err := b.client.Send(conf); err != nil {
		l.Err(err).Msg("could not send welcome message")
		return
	}
}

func (b *telegramBot) stopChat(update *tgbotapi.Update) {
	l := log.With().Str("func", "stopChat").Logger()
	if b.backend == nil {
		l.Warn().Msg("no backend is set")
		return
	}

	// -- Get the chat
	c, err := b.backend.GetChatByID(update.Message.Chat.ID)
	if err != nil {
		if err != ksb.ErrNotFound {
			l.Err(err).Msg("error while getting chat")
			return
		}

		l.Debug().Msg("chat does not exist, returning...")
		return
	}

	if err := b.backend.DeleteChat(c.Id); err != nil {
		l.Err(err).Msg("error while deleting chat on firestore")
		return
	}

	// -- Notify the user
	conf := tgbotapi.NewMessage(update.Message.Chat.ID, b.texts["messageStop"])
	conf.ReplyToMessageID = update.Message.MessageID

	if _, err := b.client.Send(conf); err != nil {
		l.Err(err).Msg("could not send stop message")
		return
	}
}
