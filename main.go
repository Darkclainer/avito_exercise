package main

import (
	"database/sql"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL
);
`

func main() {
	server := NewServer(false)

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		server.Logger.Fatal("Can not create database: ", err)
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		server.Logger.Fatal("Failed to exec schema: ", err)
	}
	server.DB = ServerDB{db}

	err = http.ListenAndServe(":9090", server)
	if err != nil {
		server.Logger.Fatal(err)
	}
}
