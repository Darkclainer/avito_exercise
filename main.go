package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/Darkclainer/avito_exercise/config"
	"github.com/Darkclainer/avito_exercise/storage"
)

func NewLogger(cfg *config.Log) (*logrus.Logger, func(), error) {
	logger := logrus.New()
	nothing := func() {}
	if cfg.Path == "stderr" {
		logger.SetOutput(os.Stderr)
		return logger, nothing, nil
	}
	logFile, err := os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return nil, nothing, err
	}
	logger.SetOutput(logFile)
	return logger, func() { logFile.Close() }, nil

}
func main() {
	viper, err := config.NewViper()
	if err != nil {
		log.Fatal("Can not initialize Viper: ", err)
	}
	cfg := config.MakeConfig(viper)

	logger, closeLog, err := NewLogger(&cfg.Log)
	if err != nil {
		log.Fatal("Can not initialize logger: ", err)
	}
	defer closeLog()

	db, err := sql.Open("sqlite3", cfg.Sqlite.Path)
	if err != nil {
		logger.Fatal("Can not create database: ", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		logger.Fatal("Can not to connect to database: ", err)
	}

	dbStorage := storage.SqlStorage{DB: db}
	dbStorage.Setup()
	server := NewServer(dbStorage, logger, false)
	logger.Debug("Server started")

	err = http.ListenAndServe(":"+cfg.Server.Port, server)
	if err != nil {
		logger.Fatal(err)
	}
}
