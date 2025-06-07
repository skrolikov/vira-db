package db

type UserRepository interface {
	CreateUser(username, passwordHash string) (string, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id string) (*User, error)
}
