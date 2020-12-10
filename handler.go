package main

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Handler handles communication with the telegram bot
type Handler interface {
}

// telegramHandler is in charge of handling communication with the telegram bot
type telegramHandler struct {
	ctx     context.Context
	client  *tgbotapi.BotAPI
	updChan tgbotapi.UpdatesChannel
	lock    sync.Mutex
}

func NewHandler(token string) Handler {
	// TODO
	return &telegramHandler{}
}
