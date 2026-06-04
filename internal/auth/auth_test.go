package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/RAG-bot/internal/db"
)

func TestLoginIssuesToken(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	userID := uuid.New()
	store := &fakeUserStore{
		user: db.User{
			ID:           userID,
			Email:        "admin@example.com",
			PasswordHash: hash,
			Role:         "admin",
		},
	}
	service, err := NewService(store, "very-secret-test-key")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	token, user, err := service.Login(context.Background(), " ADMIN@example.com ", "correct-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if user.ID != userID || user.Email != "admin@example.com" || user.Role != "admin" {
		t.Fatalf("unexpected user: %+v", user)
	}
	parsed, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if parsed.ID != userID {
		t.Fatalf("expected parsed user %s, got %s", userID, parsed.ID)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	service, err := NewService(&fakeUserStore{
		user: db.User{ID: uuid.New(), Email: "admin@example.com", PasswordHash: hash, Role: "admin"},
	}, "very-secret-test-key")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, _, err := service.Login(context.Background(), "admin@example.com", "wrong-password"); err == nil {
		t.Fatal("expected wrong password to be rejected")
	}
}

type fakeUserStore struct {
	user db.User
}

func (f *fakeUserStore) CountUsers(context.Context) (int64, error) {
	if f.user.ID == uuid.Nil {
		return 0, nil
	}
	return 1, nil
}

func (f *fakeUserStore) CreateUser(_ context.Context, arg db.CreateUserParams) (db.User, error) {
	f.user = db.User{
		ID:           uuid.New(),
		Email:        arg.Email,
		PasswordHash: arg.PasswordHash,
		Role:         arg.Role,
	}
	return f.user, nil
}

func (f *fakeUserStore) GetUserByEmail(context.Context, string) (db.User, error) {
	return f.user, nil
}

func (f *fakeUserStore) GetUserByID(context.Context, uuid.UUID) (db.User, error) {
	return f.user, nil
}
