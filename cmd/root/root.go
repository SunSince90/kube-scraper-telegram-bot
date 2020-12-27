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
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/SunSince90/kube-scraper-telegram-bot/pkg/backend/firestore"
	"github.com/SunSince90/kube-scraper-telegram-bot/pkg/bot"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	log zerolog.Logger
)

// NewRootCommand returns an instance of the telegram command
func NewRootCommand() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use: `bot -t|--token <token> -s|--service-account-path <service-account-path>
-p|--firebase-project-name <project-name> -c|--firestore-collection-name <collection-name>`,
		Example: `bot -t sg8Svd12 -s ./creds/service-account.json -p my-project -c chats`,
		Short:   "listens for chat messages",
		Long: `telegram listens for chat messages from users on telegram and sets the data
to firestore.`,
		Run: func(c *cobra.Command, args []string) {
			if opts.debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
				opts.debug = true
			}

			runTelegram(opts)
		},
	}

	// -- Flags
	cmd.Flags().StringVarP(&opts.telegramToken, "token", "t", "", "the token of the telegram bot")
	cmd.Flags().StringVarP(&opts.serviceAccountPath, "service-account-path", "s", "", "the path of gcp service account")
	cmd.Flags().StringVarP(&opts.firebaseProjectName, "firebase-project-name", "p", "", "the firebase project name")
	cmd.Flags().StringVarP(&opts.firestoreCollectionName, "firestore-collection-name", "c", "", "the name of the collection where chats are stored")
	cmd.Flags().StringVar(&opts.textsPath, "texts-path", "./texts.yaml", "the paths where to find the texts")
	cmd.Flags().BoolVarP(&opts.debug, "debug-mode", "d", false, "whether to log debug lines")

	// -- Mark as required
	cmd.MarkFlagRequired("token")
	cmd.MarkFlagRequired("service-account-path")
	cmd.MarkFlagRequired("firebase-project-name")
	cmd.MarkFlagRequired("firestore-collection-name")

	return cmd
}

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func runTelegram(opts *options) {
	l := log.With().Str("func", "runTelegram").Logger()
	l.Info().Msg("starting...")
	l.Debug().Msg("debug mode requested")

	stopChan := make(chan struct{})
	ctx, canc := context.WithCancel(context.Background())

	// -- Get the texts
	var texts map[string]string
	yamlFile, err := ioutil.ReadFile(opts.textsPath)
	if err != nil {
		l.Fatal().Err(err).Msg("error while getting texts")
	}
	err = yaml.Unmarshal(yamlFile, &texts)
	if err != nil {
		l.Fatal().Err(err).Msg("error while unmarshaling texts file")
	}

	// -- Get the backend
	fs, err := firestore.NewBackend(ctx, opts.serviceAccountPath, &firestore.Options{
		ProjectName:     opts.firebaseProjectName,
		ChatsCollection: opts.firestoreCollectionName,
		UseCache:        true,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("error while getting firestore as backend")
	}
	defer fs.Close()

	// -- Start the bot
	tgBot, err := bot.NewBotListener(&bot.TelegramOptions{
		Token: opts.telegramToken,
		Debug: &opts.debug,
	}, texts)
	if err != nil {
		l.Fatal().Err(err).Msg("could not get telegram bot")
	}

	// -- Listen for updates
	go tgBot.ListenForUpdates(ctx, stopChan)

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
	l.Info().Msg("exit requested")

	canc()
	<-stopChan

	l.Info().Msg("goodbye!")
}
