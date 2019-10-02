package main

import (
	"database/sql"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/Darkclainer/avito_exercise/storage"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER NOT NULL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL
);
`

func main() {
	logger := logrus.New()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		logger.Fatal("Can not create database: ", err)
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		logger.Fatal("Failed to exec schema: ", err)
	}
	server := NewServer(storage.SqlStorage{db}, logger, false)

	err = http.ListenAndServe(":9090", server)
	if err != nil {
		logger.Fatal(err)
	}
}
