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

func (db SqlStorage) Setup() error {
	schema := `
		CREATE TABLE IF NOT EXISTS users (
		    id INTEGER NOT NULL PRIMARY KEY,
		    username TEXT NOT NULL UNIQUE,
		    created_at DATETIME NOT NULL
		);
		CREATE TABLE IF NOT EXISTS chats (
		    id INTEGER NOT NULL PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE,
		    created_at DATETIME NOT NULL
		);
		CREATE TABLE IF NOT EXISTS users_chats (
		    user_id INTEGER NOT NULL,
		    chat_id INTEGER NOT NULL,
		    FOREIGN KEY (user_id) REFERENCES users (id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,
		    FOREIGN KEY (chat_id) REFERENCES chats (id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,
		    PRIMARY KEY (user_id, chat_id)
		);
		CREATE TABLE IF NOT EXISTS messages (
		    id INTEGER NOT NULL PRIMARY KEY,
		    chat_id INTEGER NOT NULL,
		    author_id INTEGER NOT NULL,
		    text TEXT,
		    created_at DATETIME NOT NULL,
		    FOREIGN KEY (chat_id) REFERENCES chats (id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,
		    FOREIGN KEY (author_id) REFERENCES users (id)
			ON UPDATE CASCADE
			ON DELETE CASCADE
		);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("Setup failed: %s", err)
	}
	return nil
}

func (db SqlStorage) IsUserExists(username string) (bool, error) {
	sqlStmt := `SELECT username FROM users WHERE username = ?`
	err := db.QueryRow(sqlStmt, username).Scan(&username)
	return isExistByError(err)
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
	return isExistByError(err)
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
		if _, err = db.Exec("INSERT INTO users_chats(user_id, chat_id) VALUES(?, ?)", userId, chatId); err != nil {
			return
		}
	}
	return
}
func (db SqlStorage) IsUserInChat(userId int64, chatId int64) (bool, error) {
	stmt := `SELECT user_id FROM users_chats WHERE user_id = ? AND chat_id = ?`
	err := db.QueryRow(stmt, userId, chatId).Scan(&userId)
	return isExistByError(err)
}
func (db SqlStorage) AddMessage(authorId int64, chatId int64, text string) (int64, error) {
	isUserInChat, err := db.IsUserInChat(authorId, chatId)
	if err != nil {
		return 0, err
	}
	if !isUserInChat {
		return 0, fmt.Errorf("user is not in chat, or either of them doesn't exist")
	}
	insertStatement := `INSERT INTO messages(chat_id, author_id, text, created_at) VALUES(?, ?, ?, ?)`
	result, err := db.Exec(insertStatement, chatId, authorId, text, time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func isExistByError(err error) (bool, error) {
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
