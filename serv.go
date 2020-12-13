package main

import (
	"context"

	"github.com/SunSince90/telegram-listener/listenerserv"
)

type server struct {
	mainCtx context.Context
	tg      Handler
	listenerserv.UnimplementedTelegramListenerServer
}

// NewServer returns a new instance of the grpc server
func NewServer(ctx context.Context, telh Handler) listenerserv.TelegramListenerServer {
	return &server{
		mainCtx: ctx,
		tg:      telh,
	}
}

// SendMessage uses the telegram handler to send the specified message
func (s *server) SendMessage(ctx context.Context, msg *listenerserv.TelegramMessage) (*listenerserv.TelegramResponse, error) {
	err := s.tg.SendMessage(msg.Dest, msg.Msg)
	if err != nil {
		return &listenerserv.TelegramResponse{
			Code: 500,
			Resp: err.Error(),
		}, err
	}

	return &listenerserv.TelegramResponse{
		Code: 200,
		Resp: "ok",
	}, nil
}
