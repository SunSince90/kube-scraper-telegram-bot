// Copyright © 2020 Elis Lulja
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bot

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var (
	log zerolog.Logger
)

const (
	defaultTimeout int    = 3600
	defaultOffset  int    = 0
	deafulttopic   string = "telegram-events"
)

// TelegramBot is a structure that holds the telegram bot's information
type TelegramBot struct {
	Client  *tgbotapi.BotAPI
	redis   *redis.Client
	topic   string
	updChan tgbotapi.UpdatesChannel
	log     zerolog.Logger
	lock    sync.Mutex
}

// Option represents an option for the telegram bot
type Option func(*TelegramBot)

// WithLogger sets a logger to the telegram bot
func WithLogger(z zerolog.Logger) Option {
	return func(tb *TelegramBot) {
		tb.log = z
	}
}

// WithRedisClient sets the redis client
func WithRedisClient(r *redis.Client) Option {
	return func(tb *TelegramBot) {
		tb.redis = r
	}
}

// WithTopicName sets the topic name for publishing events
func WithTopicName(topic string) Option {
	return func(tb *TelegramBot) {
		tb.topic = topic
	}
}

// NewBotListener returns a new instance of the bot listener.
func NewBotListener(token string, opts ...Option) (*TelegramBot, error) {
	// -- Get the client
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	// -- Set the struct
	locale, _ := time.LoadLocation("Europe/Rome")
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(locale)
	}
	tgBot := &TelegramBot{
		Client: bot,
		log:    zerolog.New(output).With().Timestamp().Logger(),
	}
	for _, opt := range opts {
		opt(tgBot)
	}

	// -- Get the updates channel
	u := tgbotapi.NewUpdate(defaultOffset)
	u.Timeout = defaultTimeout
	updChan, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}
	tgBot.updChan = updChan

	tgBot.log.Info().Str("account", bot.Self.UserName).Msg("authorized")
	return tgBot, nil
}

// ListenForUpdates starts an infinite loop, getting updates from the bot.
func (b *TelegramBot) ListenForUpdates(ctx context.Context) {
	var u tgbotapi.Update
	for {
		select {
		case u = <-b.updChan:
			b.parseUpdate(&u)
		case <-ctx.Done():
			return
		}
	}
}

func (b *TelegramBot) parseUpdate(update *tgbotapi.Update) {
	l := b.log.With().Int("update-id", update.UpdateID).Logger()

	if update.Message == nil {
		// ignore any non-Message Updates
		b.log.Info().Msg("got non-message update. Skipping...")
		return
	}

	l = l.With().Str("from", update.Message.From.FirstName).Int64("chat-id", update.Message.Chat.ID).Logger()
	text := update.Message.Text
	if len(text) > 200 {
		text = fmt.Sprintf("%s...", text[0:100])
	}
	l.Debug().Int64("chat-id", update.Message.Chat.ID).Str("title", update.Message.Chat.Title).
		Str("type", update.Message.Chat.Title).Str("text", text).Msg("got message")

	switch update.Message.Text {
	case "/start", "/restart":
		b.startChat(update)
	case "/stop":
		b.stopChat(update)

	// .
	// .
	// .
	// Insert other commands here...
	// .
	// .
	// .

	default:
		// Nothing is printed if the message is not recognized
		l.Info().Msg("command not recognized: nothing will be printed...")
	}
}

func (b *TelegramBot) startChat(update *tgbotapi.Update) {
	if b.redis != nil {
		b.log.Warn().Msg("redis is not set, returning...")
		return
	}

	if update.Message == nil {
		// Ignore non-message event
		return
	}

	// Handle the event, i.e. publish the event as so:
	ctx, canc := context.WithTimeout(context.Background(), 15*time.Second)
	defer canc()
	if res := b.redis.Publish(ctx, b.topic, update.Message); res.Err() != nil {
		b.log.Err(res.Err()).Msg("could not publish event, returning...")
	}
}

func (b *TelegramBot) stopChat(update *tgbotapi.Update) {
	// TODO: handle the event, i.e. as above
}
