package auth

import "testing"

func TestLoginRejectsEmptyPassword(t *testing.T) {
	service := NewAuthService(fakeUsers{})
	if _, err := service.Login("admin@example.com", ""); err != ErrInvalidCredentials {
		t.Fatalf("err = %v", err)
	}
}

type fakeUsers struct{}

func (fakeUsers) FindByEmail(email string) (User, error) {
	return User{Email: email, PasswordHash: "hash"}, nil
}
