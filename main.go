package main

import (
	"fmt"
	"os"

	"github.com/SunSince90/kube-scraper-telegram-bot/pkg/cmd/root"
	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger
)

func main() {
	cmd := root.GetRootCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
