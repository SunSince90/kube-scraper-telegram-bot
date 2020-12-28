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

package firestore

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SunSince90/kube-scraper-backend/pkg/firestore"
	"github.com/SunSince90/kube-scraper-telegram-bot/cmd/internal"
	"github.com/SunSince90/kube-scraper-telegram-bot/pkg/bot"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	log zerolog.Logger
)

// NewFirestoreCommand returns an instance of the telegram command
func NewFirestoreCommand() *cobra.Command {
	opts := &firestoreOptions{}
	cmd := &cobra.Command{
		Use: `firestore -s|--service-account-path <service-account-path>
--project-id <project-id> --chats-collection <collection-name>`,
		Example: `bot firestore --token sg8Svd12 -s ./creds/service-account.json --project-id my-project
--chats-collection chats`,
		Short: "store chat on firestore",
		Long:  `this command will use firestore as the backend.`,
		Run: func(cmd *cobra.Command, args []string) {
			topts, err := internal.GetTelegramOptions(cmd)
			if err != nil {
				log.Fatal().Err(err).Msg("could not start command")
			}

			runFirestore(opts, topts)
		},
	}

	// -- Flags
	cmd.Flags().StringVarP(&opts.serviceAccountPath, "service-account-path", "s", "", "the path to the service account file")
	cmd.Flags().StringVar(&opts.projectID, "project-id", "", "the id of the project from firebase")
	cmd.Flags().StringVar(&opts.chatsCollection, "chats-collection", "", "whether to log debug lines")

	// -- Mark as required
	cmd.MarkFlagRequired("service-account-path")
	cmd.MarkFlagRequired("project-id")
	cmd.MarkFlagRequired("chats-collection")

	return cmd
}

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func runFirestore(opts *firestoreOptions, topts *bot.TelegramOptions) {
	l := log.With().Str("func", "runFirestore").Logger()
	l.Info().Msg("starting...")
	l.Debug().Msg("debug mode requested")

	stopChan := make(chan struct{})
	ctx, canc := context.WithCancel(context.Background())

	// -- Get the backend
	fs, err := firestore.NewBackend(ctx, opts.serviceAccountPath, &firestore.Options{
		ProjectID:       opts.projectID,
		ChatsCollection: opts.chatsCollection,
		UseCache:        true,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("error while getting firestore as backend")
	}
	defer fs.Close()

	// -- Start the bot
	tgBot, err := bot.NewBotListener(topts, topts.Texts, fs)
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
