package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	db_addr = "postgres://postgres:psql123@localhost:5432/postgres?sslmode=disable"
	token   = "7192337917:AAFAu1rQkkVJZZ0Uk_5s2npMwOg5aw35uBU"
)

func main() {
	const op = "main:main"
	log.Printf("%s:%s", op, "Database initing")
	db, err := NewDatabase(db_addr)
	if err != nil {
		log.Fatalf("%s:%s", op, err)
	}
	log.Printf("%s:%s", op, "Creating migrations")
	m, err := migrate.New("file://migrations", db_addr)
	if err != nil {
		log.Fatalf("%s:%s", op, err)
	}
	log.Printf("%s:%s", op, "Migrations are appling")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("%s:%s", op, err)
	}
	server, err := NewServer(token, db)
	if err != nil {
		log.Fatalf("%s:%s", op, err)
	}

	log.Printf("%s:%s", op, "Server started!")
	server.ListAndServe()

}
