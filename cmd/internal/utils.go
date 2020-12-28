package internal

import (
	"io/ioutil"

	"github.com/SunSince90/kube-scraper-telegram-bot/pkg/bot"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// GetTelegramOptions get the options from the command
func GetTelegramOptions(cmd *cobra.Command) (*bot.TelegramOptions, error) {
	// -- Get texts
	textsPath, err := cmd.Flags().GetString("texts-path")
	if err != nil {
		return nil, err
	}
	var texts map[string]string
	yamlFile, err := ioutil.ReadFile(textsPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &texts)
	if err != nil {
		return nil, err
	}

	// -- Get the token
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return nil, err
	}

	// -- Get the debug
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}

	return &bot.TelegramOptions{
		Token: token,
		Texts: texts,
		Debug: &debug,
	}, nil
}
