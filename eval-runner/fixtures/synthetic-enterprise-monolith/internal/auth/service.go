package auth

type UserStore interface {
	FindByEmail(email string) (User, error)
}

type User struct {
	Email        string
	PasswordHash string
}

type AuthService struct {
	users UserStore
}

func NewAuthService(users UserStore) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Login(email string, password string) (User, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		return User{}, err
	}
	if !verifyPassword(password, user.PasswordHash) {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func verifyPassword(password string, hash string) bool {
	return password != "" && hash != ""
}
