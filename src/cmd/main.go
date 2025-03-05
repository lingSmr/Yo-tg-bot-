package main

import (
	"Yo/src/botServe"
	"Yo/src/config"
	"Yo/src/postgres"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	const op = "Main:main"
	config := config.InitConfig()
	defer config.LogFile.Close()
	slog.SetDefault(config.Logger)
	ctx, cancel := context.WithCancel(context.Background())

	slog.Debug(
		"Start initing",
		"Operation", op,
	)

	slog.Debug(
		"Connecting to database",
		"Operation", op,
	)
	db, err := postgres.NewPostgresDb(config.DbAddr, ctx)
	if err != nil {
		slog.Error("Cant connect to database", "Operation", op, "Error", err)
		panic(err)
	}

	server, err := botServe.NewBotServ(config.Token, db, config.Logger)
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	quit := make(chan os.Signal, 1)

	go func() {
		err = server.ListAndServe(ctx)
		if err != nil {
			slog.Error("Fatal error", "Operation", "Main:ListenAndServe", "Error", err)
			close(done)
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
		slog.Info("Take signal to shutdown")
	case <-done:
		slog.Info("Error while ListenAndServe")
	}

	slog.Debug("Shuting down")
	cancel()
	slog.Info("Programm finished")
}
