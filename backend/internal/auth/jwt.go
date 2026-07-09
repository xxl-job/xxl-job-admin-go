// Package auth JWT 签发/校验与 RBAC。
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 角色
const (
	RoleUser  int8 = 0
	RoleAdmin int8 = 1
)

// Claims JWT 载荷
type Claims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"usr"`
	Role     int8   `json:"rol"`
	jwt.RegisteredClaims
}

// JWTManager JWT 签发与校验。
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager 创建 JWT 管理器。
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl}
}

// Sign 签发 token。
func (m *JWTManager) Sign(userID uint64, username string, role int8) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

// TTL 返回 token 有效期。
func (m *JWTManager) TTL() time.Duration { return m.ttl }

// Parse 校验并解析 token。
func (m *JWTManager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
