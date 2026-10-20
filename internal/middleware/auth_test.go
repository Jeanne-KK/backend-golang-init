package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func jwtSvcForMiddleware() *jwt.Service {
	return jwt.NewService(config.JWTConfig{
		AccessSecret: "test-secret-at-least-32-characters-long",
		AccessTTL:    time.Minute,
	})
}

func TestAuthRequiredMissingHeader(t *testing.T) {
	r := gin.New()
	r.GET("/protected", AuthRequired(jwtSvcForMiddleware()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d", w.Code)
	}
	var body dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Message != "missing authorization header" {
		t.Fatalf("message: got %q", body.Message)
	}
}

func TestAuthRequiredInvalidBearerFormat(t *testing.T) {
	r := gin.New()
	r.GET("/protected", AuthRequired(jwtSvcForMiddleware()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d", w.Code)
	}
}

func TestAuthRequiredValidToken(t *testing.T) {
	svc := jwtSvcForMiddleware()
	issued, err := svc.Issue(42, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	var gotUserID uint
	var gotEmail string
	r := gin.New()
	r.GET("/protected", AuthRequired(svc), func(c *gin.Context) {
		gotUserID = c.GetUint(ContextUserIDKey)
		gotEmail = c.GetString(ContextUserEmailKey)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d body %s", w.Code, w.Body.String())
	}
	if gotUserID != 42 {
		t.Fatalf("userID: got %d", gotUserID)
	}
	if gotEmail != "user@example.com" {
		t.Fatalf("email: got %q", gotEmail)
	}
}

func TestOptionalAuthNoHeader(t *testing.T) {
	called := false
	r := gin.New()
	r.GET("/", OptionalAuth(jwtSvcForMiddleware()), func(c *gin.Context) {
		called = true
		_, ok := c.Get(ContextUserIDKey)
		if ok {
			t.Fatal("expected no user id in context")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called || w.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, w.Code)
	}
}
