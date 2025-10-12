package utils

import (
	"RIP-WEB/internal/app/ds"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Claims struct {
	UserID      uint   `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	redisClient *redis.Client
}

func NewTokenManager(redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		redisClient: redisClient,
	}
}

func GenerateJWT(user *ds.Users, secret string, expirationHours int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	claims := &Claims{
		UserID:      user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func (tm *TokenManager) AddToBlacklist(token string, expiration time.Duration) error {
	ctx := context.Background()

	// Проверяем подключение к Redis
	if err := tm.redisClient.Ping(ctx).Err(); err != nil {
		logrus.Errorf("Redis connection error in AddToBlacklist: %v", err)
		return fmt.Errorf("redis connection failed: %v", err)
	}

	// Добавляем токен в blacklist
	err := tm.redisClient.Set(ctx, "blacklist:"+token, "1", expiration).Err()
	if err != nil {
		logrus.Errorf("Failed to add token to Redis blacklist: %v", err)
		return fmt.Errorf("failed to add token to blacklist: %v", err)
	}

	// Проверяем что токен действительно добавлен
	result, err := tm.redisClient.Get(ctx, "blacklist:"+token).Result()
	if err != nil {
		logrus.Errorf("Failed to verify token in blacklist: %v", err)
	} else {
		logrus.Infof("Token successfully added to blacklist, Redis returned: %s", result)
	}

	return nil
}

func (tm *TokenManager) IsTokenBlacklisted(token string) bool {
	ctx := context.Background()

	// Сначала проверяем подключение к Redis
	if err := tm.redisClient.Ping(ctx).Err(); err != nil {
		logrus.Errorf("Redis connection error in IsTokenBlacklisted: %v", err)
		return false // Если Redis недоступен, считаем что токен не в blacklist
	}

	result, err := tm.redisClient.Get(ctx, "blacklist:"+token).Result()
	if err == nil && result == "1" {
		logrus.Infof("Token found in blacklist, rejecting request")
		return true
	}
	if err != nil && err != redis.Nil {
		logrus.Errorf("Redis error while checking blacklist: %v", err)
	}
	return false
}

// GetTokenExpiration получает оставшееся время жизни токена
func (tm *TokenManager) GetTokenExpiration(token string) (time.Duration, error) {
	ctx := context.Background()
	return tm.redisClient.TTL(ctx, "blacklist:"+token).Result()
}

// TestConnection тестирует подключение к Redis
func (tm *TokenManager) TestConnection() error {
	ctx := context.Background()
	return tm.redisClient.Ping(ctx).Err()
}

// TestSetGet тестирует запись и чтение из Redis
func (tm *TokenManager) TestSetGet(key, value string, expiration time.Duration) (string, error) {
	ctx := context.Background()

	// Запись
	err := tm.redisClient.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return "", err
	}

	// Чтение
	result, err := tm.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return result, nil
}
