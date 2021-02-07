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

package root

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SunSince90/kube-scraper-telegram-bot/pkg/bot"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	log zerolog.Logger
)

func init() {
	locale, _ := time.LoadLocation("Europe/Rome")
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(locale)
	}
	log = zerolog.New(output).With().Timestamp().Logger()
}

// GetRootCommand returns the root command
func GetRootCommand() *cobra.Command {
	opts := &options{
		redis: &redisOptions{},
	}

	cmd := &cobra.Command{
		Use:     `bot --token <token> [--texts-path <texts-path> --debug]`,
		Example: `bot --token sg8Svd12 `,
		Short:   "listen for chat messages",
		Long:    `telegram listens for chat messages from users on telegram.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opts.debugMode {
				log = log.With().Logger().Level(zerolog.DebugLevel)
			}
		},
		Run: func(_ *cobra.Command, _ []string) {
			run(opts)
		},
	}

	// -- Flags
	cmd.Flags().StringVar(&opts.token, "token", "", "the token of the telegram bot")
	cmd.Flags().BoolVar(&opts.debugMode, "debug", false, "whether to log debug lines")

	cmd.Flags().StringVar(&opts.redis.address, "redis-address", "", "the address of redis service")
	cmd.Flags().StringVar(&opts.redis.topic, "redis-topic", "telegram-event", "the name of the topic where to publish events")

	// -- Mark as required
	cmd.MarkFlagRequired("token")
	cmd.MarkFlagRequired("redis-address")

	return cmd
}

func run(opts *options) {
	rdb := redis.NewClient(&redis.Options{Addr: opts.redis.address})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatal().Msg("could not get pong message")
		return
	}
	defer rdb.Close()

	tgbot, err := bot.NewBotListener(opts.token, bot.WithLogger(log), bot.WithRedisClient(rdb))
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while starting bot")
		return
	}

	ctx, canc := context.WithCancel(context.Background())
	exitChan := make(chan struct{})

	go func() {
		tgbot.ListenForUpdates(ctx)
		close(exitChan)
	}()

	// -- Graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	fmt.Println()
	log.Info().Msg("exit requested")

	canc()
	<-exitChan

	log.Info().Msg("goodbye!")
}
