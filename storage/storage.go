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
	GetMessagesFromChat(chatId int64) ([]*Message, error)
}

type Chat struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UserIds   []int64   `json:"users"`
}

type Message struct {
	Id        int64     `json:"id"`
	ChatId    int64     `json:"chat"`
	AuthorId  int64     `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
