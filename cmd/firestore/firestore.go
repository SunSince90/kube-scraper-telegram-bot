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
	"os"

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
			runFirestore(opts)
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

func runFirestore(opts *firestoreOptions) {
	// TODO: implement me
}
