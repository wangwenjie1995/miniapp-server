package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager([]byte("test-secret-at-least-32-bytes-long"), time.Hour)
	raw, expires, err := manager.Generate(42)
	if err != nil {
		t.Fatal(err)
	}
	id, err := manager.Verify(raw)
	if err != nil || id != 42 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	claims := &jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return manager.secret, nil }); err != nil {
		t.Fatal(err)
	}
	if !expires.Equal(claims.ExpiresAt.Time) {
		t.Fatal("response expiration differs from JWT")
	}
}

func TestRejectInvalidTokens(t *testing.T) {
	secret := []byte("test-secret-at-least-32-bytes-long")
	manager := NewTokenManager(secret, time.Hour)
	for _, tc := range []struct {
		name, subject string
		exp           *jwt.NumericDate
		method        jwt.SigningMethod
		key           []byte
	}{
		{"expired", "1", jwt.NewNumericDate(time.Now().Add(-time.Hour)), jwt.SigningMethodHS256, secret},
		{"missing expiration", "1", nil, jwt.SigningMethodHS256, secret},
		{"wrong signature", "1", jwt.NewNumericDate(time.Now().Add(time.Hour)), jwt.SigningMethodHS256, []byte("different-secret-at-least-32-bytes")},
		{"wrong algorithm", "1", jwt.NewNumericDate(time.Now().Add(time.Hour)), jwt.SigningMethodHS384, secret},
		{"zero subject", "0", jwt.NewNumericDate(time.Now().Add(time.Hour)), jwt.SigningMethodHS256, secret},
		{"invalid subject", "abc", jwt.NewNumericDate(time.Now().Add(time.Hour)), jwt.SigningMethodHS256, secret},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := jwt.NewWithClaims(tc.method, jwt.RegisteredClaims{Subject: tc.subject, ExpiresAt: tc.exp}).SignedString(tc.key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := manager.Verify(raw); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
	if _, err := manager.Verify("invalid"); err == nil {
		t.Fatal("malformed token accepted")
	}
}
