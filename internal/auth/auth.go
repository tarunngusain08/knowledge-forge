package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/tarunngusain08/RAG-bot/internal/db"
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

type Service struct {
	store     db.Querier
	jwtSecret []byte
	now       func() time.Time
}

func NewService(store db.Querier, jwtSecret string) (*Service, error) {
	if store == nil {
		return nil, errors.New("auth store is required")
	}
	if len(jwtSecret) < 12 {
		return nil, errors.New("JWT_SECRET must be at least 12 characters")
	}
	return &Service{
		store:     store,
		jwtSecret: []byte(jwtSecret),
		now:       time.Now,
	}, nil
}

func (s *Service) SeedAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	count, err := s.store.CountUsers(ctx)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		return nil
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.store.CreateUser(ctx, db.CreateUserParams{
		Email:        normalizeEmail(email),
		PasswordHash: hash,
		Role:         "admin",
	})
	if err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, User, error) {
	row, err := s.store.GetUserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		return "", User{}, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return "", User{}, errors.New("invalid email or password")
	}
	user := User{ID: row.ID, Email: row.Email, Role: row.Role}
	token, err := s.IssueToken(user, 24*time.Hour)
	if err != nil {
		return "", User{}, err
	}
	return token, user, nil
}

func (s *Service) IssueToken(user User, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(s.now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(s.now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(tokenValue string) (User, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenValue, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return User{}, errors.New("invalid token")
	}
	return User{ID: claims.UserID, Email: claims.Email, Role: claims.Role}, nil
}

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}
