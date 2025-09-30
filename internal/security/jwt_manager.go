package security

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AccessTokenManager interface {
	GenerateAccessToken(userID int64, username string) (string, time.Time, error)
	ValidateAccessToken(token string) (*AccessTokenClaims, error)
}

type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

func NewJWTManager(secret, issuer, audience string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("access token secret must be at least 32 characters")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("access token ttl must be positive")
	}
	return &JWTManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
		now:      time.Now,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(userID int64, username string) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.ttl)
	subject := strconv.FormatInt(userID, 10)

	claims := AccessTokenClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
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
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *JWTManager) ValidateAccessToken(token string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(5 * time.Second),
	}
	if m.issuer != "" {
		opts = append(opts, jwt.WithIssuer(m.issuer))
	}
	if m.audience != "" {
		opts = append(opts, jwt.WithAudience(m.audience))
	}

	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	}, opts...)
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("token invalid")
	}
	if claims.ExpiresAt == nil || time.Now().UTC().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

func (m *JWTManager) WithNow(now func() time.Time) {
	if now != nil {
		m.now = now
	}
}
