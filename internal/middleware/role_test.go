package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/model"
	"backend/internal/pkg/jwt"
	"backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type roleTestUserRepo struct {
	byID map[uint]*model.User
}

func (r *roleTestUserRepo) Create(context.Context, *model.User) error { return nil }
func (r *roleTestUserRepo) FindByEmail(context.Context, string) (*model.User, error) {
	return nil, nil
}
func (r *roleTestUserRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
	return r.byID[id], nil
}
func (r *roleTestUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	return false, nil
}
func (r *roleTestUserRepo) FindAll(context.Context) (*[]model.User, error) {
	return nil, nil
}

func TestRequireRoleWithoutAuthContext(t *testing.T) {
	r := gin.New()
	repo := &roleTestUserRepo{byID: map[uint]*model.User{}}
	r.GET("/admin", RequireRole(repo, model.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d", w.Code)
	}
}

func TestRequireRoleForbiddenForUserRole(t *testing.T) {
	repo := &roleTestUserRepo{
		byID: map[uint]*model.User{
			1: {ID: 1, Role: model.RoleUser},
		},
	}

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set(ContextUserIDKey, uint(1))
		c.Next()
	}, RequireRole(repo, model.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status: got %d body %s", w.Code, w.Body.String())
	}
	var body dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "forbidden" {
		t.Fatalf("code: got %q", body.Code)
	}
}

func TestRequireRoleAllowsAdmin(t *testing.T) {
	repo := &roleTestUserRepo{
		byID: map[uint]*model.User{
			7: {ID: 7, Role: model.RoleAdmin},
		},
	}

	var gotRole string
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set(ContextUserIDKey, uint(7))
		c.Next()
	}, RequireRole(repo, model.RoleAdmin), func(c *gin.Context) {
		gotRole = c.GetString(ContextUserRoleKey)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d", w.Code)
	}
	if gotRole != model.RoleAdmin {
		t.Fatalf("role: got %q", gotRole)
	}
}

func TestRequireRoleWithAuthRequiredChain(t *testing.T) {
	svc := jwt.NewService(config.JWTConfig{
		AccessSecret: "test-secret-at-least-32-characters-long",
		AccessTTL:    time.Minute,
	})
	issued, err := svc.Issue(9, "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}

	repo := &roleTestUserRepo{
		byID: map[uint]*model.User{
			9: {ID: 9, Role: model.RoleAdmin},
		},
	}

	r := gin.New()
	r.GET("/admin", AuthRequired(svc), RequireRole(repo, model.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+issued.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d body %s", w.Code, w.Body.String())
	}
}

// compile-time check that roleTestUserRepo satisfies repository.UserRepository
var _ repository.UserRepository = (*roleTestUserRepo)(nil)
