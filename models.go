package db

type UserRepository interface {
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	CreateUserExtended(username, passwordHash, email, role string, confirmed bool, confirmToken string) (string, error)

	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}
