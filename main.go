package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/SunSince90/telegram-bot-listener/listenerserv"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
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
	var firebaseServAcc string
	var firebaseProjectName string
	var port int
	var address string

	// -- Parse flags
	flag.StringVarP(&token, "token", "t", "", "the telegram token")
	flag.BoolVarP(&debugMode, "debug", "d", false, "whether to log debug log lines")
	flag.IntVarP(&offset, "offset", "o", 0, "the offset to start")
	flag.IntVar(&timeout, "timeout", 3600, "timeout in listening for updates")
	flag.StringVarP(&firebaseServAcc, "firebase-service-account", "s", "", "the firebase service account")
	flag.StringVarP(&firebaseProjectName, "firebase-project", "p", "", "the firebase project id")
	flag.StringVarP(&address, "address", "a", "localhost", "the address where to listen from")
	flag.IntVar(&port, "port", 80, "the port where to listen from")
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

	if len(firebaseProjectName) == 0 {
		l.Fatal("no firebase project name provided. Exiting...")
		return
	}

	if len(firebaseServAcc) == 0 {
		l.Fatal("no firebase service account path provided. Exiting...")
		return
	}

	// Contexts and exit channels
	ctx, canc := context.WithCancel(context.Background())
	exitChan := make(chan struct{})
	l.Info("starting....")

	// -- Get the firebase client
	fs, err := NewFSHandler(ctx, firebaseProjectName, firebaseServAcc)
	if err != nil {
		l.WithError(err).Fatal("error while loading firestore")
	}
	defer fs.Close()
	l.Info("firestore client loaded successfully")

	// -- Get the handler
	h, err := NewHandler(ctx, token, offset, timeout, debugMode, fs)
	if err != nil {
		l.WithError(err).Fatal("error while loading handler")
	}

	go h.ListenForUpdates(ctx, exitChan)

	// Start the server
	serv := NewServer(ctx, h)
	endpoint := fmt.Sprintf("%s:%d", address, port)
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		l.WithError(err).Error("failed to listen")
		return
	}
	grpcServer := grpc.NewServer()
	listenerserv.RegisterTelegramListenerServer(grpcServer, serv)
	go grpcServer.Serve(lis)

	l.WithField("endpoint", endpoint).Info("serving requests...")

	// Graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	fmt.Println()
	l.Info("exit requested")

	canc()
	grpcServer.GracefulStop()
	<-exitChan

	l.Info("goodbye!")
}

func getFirebaseClient(ctx context.Context, projectName, servAcc string) (fsClient *firestore.Client, err error) {
	conf := &firebase.Config{ProjectID: projectName}
	app, err := firebase.NewApp(ctx, conf, option.WithServiceAccountFile(servAcc))
	if err != nil {

		return
	}
	fsClient, err = app.Firestore(ctx)
	if err != nil {
		return
	}

	return
}
