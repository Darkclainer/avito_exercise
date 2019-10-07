package storage

import (
	"time"
)

type Storage interface {
	IsUserExists(username string) (bool, error)
	AreUsersExistByIds(userIds []int64) (bool, error)
	AddUser(username string) (int64, error)
	GetUserChats(userId int64) ([]*Chat, error)

	IsChatExists(chatname string) (bool, error)
	AddChat(chatname string, userIds []int64) (int64, error)
	IsUserInChat(userId int64, chatId int64) (bool, error)

	AddMessage(chatId int64, authorId int64, text string) (int64, error)
}

type Chat struct {
	Id        int64
	Name      string
	CreatedAt time.Time
	UserIds   []int64
}
