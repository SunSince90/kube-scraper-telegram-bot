package main

import (
	"context"

	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	// -- Init
	var token string
	var debugMode bool
	var offset int
	var timeout int

	// -- Parse flags
	flag.StringVarP(&token, "token", "t", "", "the telegram token")
	flag.BoolVarP(&debugMode, "debug", "d", false, "whether to log debug log lines")
	flag.IntVarP(&offset, "offset", "o", 0, "the offset to start")
	flag.IntVar(&timeout, "timeout", 3600, "timeout in listening for updates")
	flag.Parse()

	// Contexts and exit channels
	ctx, canc := context.WithCancel(context.Background())
	exitChan := make(chan struct{})

	// -- Set log level
	if !debugMode {
		log.SetLevel(logrus.InfoLevel)
	}
	l := log.WithField("func", "main")

	// -- Parse flags
	if len(token) == 0 {
		l.Fatal("no token provided. Exiting...")
		return
	}

	l.Info("starting....")

	// -- Get the handler
	h, err := NewHandler(token, offset, timeout, debugMode)
	if err != nil {
		l.WithError(err).Fatal("error while loading handler")
	}

	go h.ListenForUpdates(ctx, exitChan)

	_ = ctx
	canc()
}
