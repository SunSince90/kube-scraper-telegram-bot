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
	"os"

	"github.com/SunSince90/kube-scraper-telegram-bot/cmd/firestore"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	log zerolog.Logger
)

// NewRootCommand returns an instance of the telegram command
func NewRootCommand() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:     `bot --token <token> [--texts-paths <texts-path> --debug]`,
		Example: `bot --token sg8Svd12 <backend>`,
		Short:   "listens for chat messages",
		Long:    `telegram listens for chat messages from users on telegram.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			textsPath, err := cmd.Flags().GetString("texts-path")
			if err != nil {
				log.Fatal().Msg("could not get texts-paths flag")
			}
			if _, err := os.Stat(textsPath); os.IsNotExist(err) {
				log.Fatal().Str("texts-path", textsPath).Err(err).Msg("could not find texts files")
			}
		},
	}

	// -- Flags
	cmd.PersistentFlags().String("token", "", "the token of the telegram bot")
	cmd.PersistentFlags().String("texts-path", "./texts.yaml", "the paths where to find the texts")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "whether to log debug lines")

	// -- Mark as required
	cmd.MarkPersistentFlagRequired("token")

	// -- Subcommands
	cmd.AddCommand(firestore.NewFirestoreCommand())

	return cmd
}

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
