package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenClaims struct {
	UserID string `json:"user_id"`
	IP     string `json:"ip"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secretKey []byte
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{secretKey: []byte(secret)}
}

func (s *TokenService) GenerateAccessToken(userID string, ip net.IP) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		IP:     ip.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(s.secretKey)
}

func (s *TokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *TokenService) ParseAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
