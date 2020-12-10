package main

import (
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

	// -- Parse flags
	flag.StringVarP(&token, "token", "t", "", "the telegram token")
	flag.BoolVarP(&debugMode, "debug", "d", false, "whether to log debug log lines")
	flag.Parse()

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
	// TODO: get a handler
}
