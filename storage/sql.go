package storage

import (
	"database/sql"
	"fmt"
	"strings"
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
func (db SqlStorage) AreUsersExistByIds(userIds []int64) (bool, error) {
	inStmtPart := strings.Repeat("?, ", len(userIds))
	sqlStmt := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE id IN (%s)", inStmtPart[:len(inStmtPart)-2])
	fmt.Println("STATEMENT: ", sqlStmt)
	args := make([]interface{}, len(userIds))
	for i, id := range userIds {
		args[i] = id
	}
	var numberOfRows int
	err := db.QueryRow(sqlStmt, args...).Scan(&numberOfRows)
	if err != nil {
		return false, err
	}
	return numberOfRows == len(userIds), nil
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

func (db SqlStorage) IsChatExists(chatname string) (bool, error) {
	sqlStmt := `SELECT name FROM chats WHERE name = ?`
	err := db.QueryRow(sqlStmt, chatname).Scan(&chatname)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
func (db SqlStorage) AddChat(chatname string, userIds []int64) (int64, error) {
	return int64(0), fmt.Errorf("Not implemented!")
}
