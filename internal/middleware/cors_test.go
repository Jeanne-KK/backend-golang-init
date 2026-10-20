package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"

	"github.com/gin-gonic/gin"
)

func corsCfg() config.CORSConfig {
	return config.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORS(corsCfg()))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("ACAO: %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected Allow-Credentials true")
	}
}

func TestCORSUnknownOriginNoACAO(t *testing.T) {
	r := gin.New()
	r.Use(CORS(corsCfg()))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unexpected ACAO: %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
}

func TestCORSOptionsPreflight(t *testing.T) {
	after := false
	r := gin.New()
	r.Use(CORS(corsCfg()))
	r.GET("/x", func(c *gin.Context) {
		after = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status: got %d", w.Code)
	}
	if after {
		t.Fatal("handler should not run on OPTIONS")
	}
}

func TestValidateOriginAllowed(t *testing.T) {
	r := gin.New()
	r.POST("/refresh", ValidateOrigin(corsCfg()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
}

func TestValidateOriginForbidden(t *testing.T) {
	r := gin.New()
	r.POST("/refresh", ValidateOrigin(corsCfg()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status: got %d", w.Code)
	}
}
