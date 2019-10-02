package storage

import (
	"database/sql"
	"time"
)

type SqlStorage struct {
	*sql.DB
}

func (db SqlStorage) IsUserExists(username string) (bool, error) {
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

func (db SqlStorage) AddUser(username string) (int64, error) {
	result, err := db.Exec("INSERT INTO users(username, created_at) VALUES(?, ?)",
		username,
		time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
