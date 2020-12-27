package bot

import (
	"context"
	"fmt"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/rs/zerolog"
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
	lock    sync.Mutex
}

// NewBotListener returns a new instance of the bot listener
func NewBotListener(opts *TelegramOptions, texts map[string]string) (Bot, error) {
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
		texts:   texts,
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
	// TODO: implement me
}

func (b *telegramBot) stopChat(update *tgbotapi.Update) {
	// TODO: implement me
}
