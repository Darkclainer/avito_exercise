package storage

type Storage interface {
	IsUserExists(username string) (bool, error)
	AddUser(username string) (int64, error)
}
