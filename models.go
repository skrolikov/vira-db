package db

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
	CreateUserExtended(username, passwordHash, email, role string, confirmed bool, confirmToken string) (string, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
	ConfirmUser(email, token string) error
	UpdatePassword(id, newHash string) error
	GetUsersByRole(role string, limit, offset int) ([]*User, error)
}
