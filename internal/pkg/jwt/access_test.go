package jwt

import (
	"testing"
	"time"

	"backend/internal/config"
)

func testJWTConfig(ttl time.Duration) config.JWTConfig {
	return config.JWTConfig{
		AccessSecret: "test-secret-at-least-32-characters-long",
		AccessTTL:    ttl,
	}
}

func TestIssueParseRoundTrip(t *testing.T) {
	svc := NewService(testJWTConfig(time.Minute))
	issued, err := svc.Issue(1, "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.Parse(issued.Token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Email != "a@b.com" {
		t.Fatalf("email: got %q", claims.Email)
	}
	if claims.Subject != "1" {
		t.Fatalf("subject: got %q", claims.Subject)
	}
	userID, err := UserIDFromClaims(claims)
	if err != nil {
		t.Fatal(err)
	}
	if userID != 1 {
		t.Fatalf("user id: got %d", userID)
	}
}

func TestParseWrongSecret(t *testing.T) {
	svc := NewService(testJWTConfig(time.Minute))
	issued, err := svc.Issue(1, "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	other := NewService(config.JWTConfig{
		AccessSecret: "other-secret-at-least-32-characters-long",
		AccessTTL:    time.Minute,
	})
	_, err = other.Parse(issued.Token)
	if err == nil {
		t.Fatal("expected parse error with wrong secret")
	}
}

func TestParseExpiredToken(t *testing.T) {
	svc := NewService(testJWTConfig(-time.Second))
	issued, err := svc.Issue(1, "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Parse(issued.Token)
	if err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestUserIDFromClaimsInvalidSubject(t *testing.T) {
	claims := &AccessClaims{}
	claims.Subject = "not-a-number"
	_, err := UserIDFromClaims(claims)
	if err == nil {
		t.Fatal("expected error for invalid subject")
	}
}
