package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/dto"

	"github.com/gin-gonic/gin"
)

func testHandlerConfig() *config.Config {
	return &config.Config{
		Cookie: config.CookieConfig{
			RefreshName: "refresh_token",
			Secure:      false,
			SameSite:    "Lax",
			Path:        "/api/auth/refresh",
		},
	}
}

func TestRefreshMissingCookie(t *testing.T) {
	r := gin.New()
	h := NewAuthHandler(nil, testHandlerConfig())
	r.POST("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d body %s", w.Code, w.Body.String())
	}
	var body dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Message != "missing refresh token" {
		t.Fatalf("message: got %q", body.Message)
	}
}

func TestRegisterValidationEmptyBody(t *testing.T) {
	body := runRegisterValidation(t, "{}")
	if body.Code != "validation_error" {
		t.Fatalf("code: %q", body.Code)
	}
	if len(body.Details) == 0 {
		t.Fatal("expected field details")
	}
}

func TestRegisterValidationBadEmail(t *testing.T) {
	payload := `{"name":"Ab","email":"not-email","password":"password123"}`
	body := runRegisterValidation(t, payload)
	if body.Details["email"] == "" {
		t.Fatalf("expected email detail, got %v", body.Details)
	}
}

func TestRegisterValidationShortPassword(t *testing.T) {
	payload := `{"name":"Ab","email":"a@b.com","password":"short"}`
	body := runRegisterValidation(t, payload)
	if body.Details["password"] == "" {
		t.Fatalf("expected password detail, got %v", body.Details)
	}
}

func runRegisterValidation(t *testing.T, jsonBody string) dto.ErrorResponse {
	r := gin.New()
	h := NewAuthHandler(nil, nil)
	r.POST("/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d body %s", w.Code, w.Body.String())
	}
	var body dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body
}
