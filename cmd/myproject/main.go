package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"sql-index-simulator/internal/api"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("starting sql-index-simulator application")

	// Start server
	server := api.NewServer()

	// graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			logrus.Fatalf("failed to start server: %v", err)
		}
	}()

	<-quit
	logrus.Info("shutting down sql-index-simulator application")
	if err := server.Stop(); err != nil {
		logrus.Errorf("error during server shutdown: %v", err)
	}
}


