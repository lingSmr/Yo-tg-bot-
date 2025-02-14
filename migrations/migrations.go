package migrations

import (
	"Yo/configs"
	"log/slog"

	"github.com/golang-migrate/migrate"
)

func UpMigrations() {
	const op = "migrations:main"
	config := configs.InitConfig()
	defer config.LogFile.Close()
	slog.SetDefault(config.Logger)

	m, err := migrate.New("file://migrations", config.DbAddr)
	if err != nil {
		slog.Error("Cant find migrations", "Operation", op, "Error", err)
		panic(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		slog.Error("Cant aply migrations", "Operation", op, "Error", err)
		return
	}
}
