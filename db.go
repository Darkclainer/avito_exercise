package main

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type ServerDB struct {
	*sql.DB
}

func (db ServerDB) IsUserExists(username string) (bool, error) {
	sqlStmt := `SELECT username FROM users WHERE username = ?`
	err := db.QueryRow(sqlStmt, username).Scan(&username)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (db ServerDB) AddUser(username string) (driver.Result, error) {
	result, err := db.Exec("INSERT INTO users(username, created_at) VALUES(?, ?)",
		username,
		time.Now())
	return result, err
}
