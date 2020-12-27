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
	timeout int = 3600
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

type telegramBot struct {
	client  *tgbotapi.BotAPI
	updChan tgbotapi.UpdatesChannel
	lock    sync.Mutex
}

// NewBotListener returns a new instance of the bot listener
func NewBotListener(token string, offset int, debugMode bool) (Bot, error) {
	if len(token) == 0 {
		return nil, fmt.Errorf("no token provided")
	}

	l := log.With().Str("func", "NewBotListener").Logger()

	// -- Get the client
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot.Debug = debugMode

	l.Debug().Str("account", bot.Self.UserName).Msg("authorized")

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
	}

	return b, nil
}

func (b *telegramBot) ListenForUpdates(ctx context.Context, exitChan chan struct{}) {
	// TODO: implement me
}
