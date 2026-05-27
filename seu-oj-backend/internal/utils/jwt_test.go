package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	token, err := GenerateToken("secret", 42, "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
		t.Fatalf("expected future expiration, got %+v", claims.ExpiresAt)
	}
}

func TestParseTokenRejectsInvalidSecretAndExpiredClaims(t *testing.T) {
	token, err := GenerateToken("secret", 7, "student")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err := ParseToken("other-secret", token); err == nil {
		t.Fatal("expected invalid signature error")
	}

	expiredClaims := CustomClaims{
		UserID: 8,
		Role:   "teacher",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		},
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	if _, err := ParseToken("secret", expired); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestParseTokenRejectsMalformedInput(t *testing.T) {
	if _, err := ParseToken("secret", "not-a-token"); err == nil {
		t.Fatal("expected malformed token error")
	}
}
