package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

// Инициирует .env файл , запускается до main
func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}
}

// Создает cЛоггер для stdOut и для file.log
func initLogger() (*slog.Logger, *os.File) {
	logFileAddr, ok := os.LookupEnv("LOG_FILE")
	if !ok {
		panic("No LOG_FILE in .env")
	}
	file, err := os.OpenFile(logFileAddr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("No .log file in directory")
		panic(err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	slogger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(slogger)
	return slogger, file
}

func main() {
	const op = "Main:main"
	slogger, file := initLogger()
	defer file.Close()
	slog.SetDefault(slogger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	slogger.Debug(
		"Start initing",
		"Operation", op,
	)
	token, ok := os.LookupEnv("TOKEN")
	if !ok {
		panic("No TOKEN in .env")
	}
	db_addr, ok := os.LookupEnv("POSTGRES_ADDR")
	if !ok {
		panic("No DB_ADDR in .env")
	}
	slogger.Debug(
		"Connecting to database",
		"Operation", op,
	)
	db, err := NewDatabase(db_addr, ctx)
	if err != nil {
		slogger.Error("Cant connect to database", "Operation", op, "Error", err)
		panic(err)
	}

	m, err := migrate.New("file://migrations", db_addr)
	if err != nil {
		slogger.Error("Cant find migrations", "Operation", op, "Error", err)
		panic(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		slogger.Error("Cant aply migrations", "Operation", op, "Error", err)
		return
	}

	server := NewServer(token, db, slogger)

	done := make(chan struct{})
	quit := make(chan os.Signal, 1)

	go func() {
		err = server.ListAndServe(ctx)
		if err != nil {
			slogger.Error("Fatal error", "Operation", "Main:ListenAndServe", "Error", err)
			close(done)
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Блокировка до получения сигнала
	select {
	case <-quit:
		slog.Info("Take signal to shutdown")
	case <-done:
		slog.Info("Error while ListenAndServe")
	}

	slog.Info("Shuting down")
	cancel()
	slog.Info("Programm finished")
}
