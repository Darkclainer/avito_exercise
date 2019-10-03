package main

import (
	"database/sql"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/Darkclainer/avito_exercise/storage"
)

func main() {
	logger := logrus.New()
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		logger.Fatal("Can not create database: ", err)
	}
	defer db.Close()

	dbStorage := storage.SqlStorage{db}
	dbStorage.Setup()
	server := NewServer(dbStorage, logger, false)

	err = http.ListenAndServe(":9090", server)
	if err != nil {
		logger.Fatal(err)
	}
}
