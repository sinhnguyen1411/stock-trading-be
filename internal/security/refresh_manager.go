package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshTokenClaims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type RefreshTokenManager interface {
	GenerateRefreshToken(userID int64, username string) (string, time.Time, error)
	ValidateRefreshToken(token string) (*RefreshTokenClaims, error)
	RevokeRefreshToken(token string) error
}

type JWTRefreshManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time

	mu      sync.RWMutex
	revoked map[string]time.Time
}

func NewJWTRefreshManager(secret, issuer, audience string, ttl time.Duration) (*JWTRefreshManager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("refresh token secret must be at least 32 characters")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("refresh token ttl must be positive")
	}
	return &JWTRefreshManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
		now:      time.Now,
		revoked:  make(map[string]time.Time),
	}, nil
}

func (m *JWTRefreshManager) cleanupLocked(now time.Time) {
	for id, exp := range m.revoked {
		if !exp.IsZero() && now.After(exp) {
			delete(m.revoked, id)
		}
	}
}

func (m *JWTRefreshManager) GenerateRefreshToken(userID int64, username string) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.ttl)
	id, err := randomTokenID()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate token id: %w", err)
	}
	claims := RefreshTokenClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	if m.audience != "" {
		claims.Audience = jwt.ClaimStrings{m.audience}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *JWTRefreshManager) ValidateRefreshToken(token string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("token invalid")
	}
	if claims.ExpiresAt == nil || time.Now().UTC().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}
	id := claims.ID
	m.mu.RLock()
	revokedAt, revoked := m.revoked[id]
	m.mu.RUnlock()
	if revoked {
		if revokedAt.IsZero() || time.Now().UTC().Before(revokedAt) {
			return nil, errors.New("token revoked")
		}
	}
	return claims, nil
}

func (m *JWTRefreshManager) RevokeRefreshToken(token string) error {
	claims := &RefreshTokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithLeeway(5*time.Second))
	if err != nil {
		return err
	}
	expires := time.Time{}
	if claims.ExpiresAt != nil {
		expires = claims.ExpiresAt.Time
	}
	now := m.now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revoked[claims.ID] = expires
	m.cleanupLocked(now)
	return nil
}

func (m *JWTRefreshManager) WithNow(now func() time.Time) {
	if now != nil {
		m.now = now
	}
}

func randomTokenID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
