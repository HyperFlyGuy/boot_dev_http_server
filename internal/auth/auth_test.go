package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	duration := time.Hour

	token, err := MakeJWT(userID, secret, duration)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gotID != userID {
		t.Errorf("expected %v, got %v", userID, gotID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, -time.Hour) // already expired
	if err != nil {
		t.Fatalf("expected no error making token, got %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected an error for expired token, got nil")
	}
}

func TestWrongSecretJWT(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("expected no error making token, got %v", err)
	}

	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected an error for wrong secret, got nil")
	}
}
