package main

import (
	"database/sql"
	"fmt"
	"github.com/safe-area/user-data-collector/config"
	"github.com/safe-area/user-data-collector/internal/api"
	"github.com/safe-area/user-data-collector/internal/repository"
	"github.com/safe-area/user-data-collector/internal/service"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.ParseConfig("./config/config.json", "./secrets/")
	if err != nil {
		logrus.Fatalf("parse config error: %v", err)
	}

	pgConn, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PgConfig.Host,
		cfg.PgConfig.Port,
		cfg.PgConfig.User,
		cfg.PgConfig.Pass,
		cfg.PgConfig.DB,
	))
	if err != nil {
		logrus.Fatalf("postgres connection error: %v", err)
	}

	repo := repository.New(pgConn)

	svc := service.New(cfg, repo)

	server := api.New(svc, cfg.ServerPort)

	errChan := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		errChan <- server.Start()
	}()

	logrus.Info("server started")

	select {
	case err := <-errChan:
		if err != nil {
			logrus.Errorf("server crushed with error: %v", err)
		}
		server.Shutdown()
	case <-signalChan:
		logrus.Info("received a signal, shutting down...")
		server.Shutdown()
	}
}
