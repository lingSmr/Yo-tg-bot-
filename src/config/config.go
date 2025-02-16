package config

import (
	"io"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token   string
	DbAddr  string
	Logger  *slog.Logger
	LogFile *os.File
}

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}
}

func InitConfig() *Config {
	logUrl, ok := os.LookupEnv("LOG_URL")
	if !ok {
		slog.Error("No LOG_URL in .env file")
		panic("Need log file address")
	}
	slogger, logFile := initLogger(logUrl)

	slog.SetDefault(slogger)
	token, ok := os.LookupEnv("TOKEN")
	if !ok {
		slog.Error("No TOKEN in .env file")
		panic("Need token")
	}
	dbAddr, ok := os.LookupEnv("POSTGRES_ADDR")
	if !ok {
		slog.Error("No POSTGRES_ADDR in .env file")
		panic("Need database address")
	}

	return &Config{Token: token, DbAddr: dbAddr, Logger: slogger, LogFile: logFile}
}

// Создает cЛоггер для stdOut и для file.log
func initLogger(logUrl string) (*slog.Logger, *os.File) {
	file, err := os.OpenFile(logUrl, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("No .log file in directory")
		panic(err)
	}
	multiWriter := io.MultiWriter(file, os.Stdout)
	slogger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(slogger)
	return slogger, file
}

func GetLogUrl() string {
	logUrl, _ := os.LookupEnv("LOG_URL")
	return logUrl

}
