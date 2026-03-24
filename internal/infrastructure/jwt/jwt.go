package jwt

import (
	"errors"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenTTL = 24 * time.Hour

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Role   entity.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
}

// NewTokenManager создаёт менеджер JWT-токенов с заданным секретом.
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

// GenerateToken выпускает подписанный JWT с идентификатором пользователя и ролью.
func (m *TokenManager) GenerateToken(userID uuid.UUID, role entity.Role) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken проверяет подпись и срок действия токена, возвращает claims при успехе.
func (m *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
