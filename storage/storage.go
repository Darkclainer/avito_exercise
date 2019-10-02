package storage

type Storage interface {
	IsUserExists(username string) (bool, error)
	AreUsersExistByIds(userIds []int64) (bool, error)
	AddUser(username string) (int64, error)

	IsChatExists(chatname string) (bool, error)
	AddChat(chatname string, userIds []int64) (int64, error)
}
