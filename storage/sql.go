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
func (db SqlStorage) AddChat(chatName string, userIds []int64) (chatId int64, err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()
	result, err := db.Exec("INSERT INTO chats(name, created_at) VALUES(?, ?)", chatName, time.Now())
	if err != nil {
		return
	}
	if chatId, err = result.LastInsertId(); err != nil {
		return
	}
	for _, userId := range userIds {
		if _, err = db.Exec("INSERT INTO users_chats(user_id, chat_id) VALUES(?, ?)", chatId, userId); err != nil {
			return
		}
	}
	return
}
