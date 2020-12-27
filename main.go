package main

import (
	"fmt"
	"os"

	"github.com/SunSince90/kube-scraper-telegram-bot/cmd/root"
	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger
)

func main() {
	cmd := root.NewRootCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
