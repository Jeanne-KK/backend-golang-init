package main

import (
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/password"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.OpenMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("mysql connect: %v", err)
	}

	hasher := password.NewHasher(cfg.Argon2)
	jwtSvc := jwt.NewService(cfg.JWT)

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	txMgr := repository.NewTxManager(db)
	authSvc := service.NewAuthService(cfg, userRepo, refreshRepo, txMgr, hasher, jwtSvc)
	userSvc := service.NewUserService(userRepo)

	authH := handler.NewAuthHandler(authSvc, cfg)
	userH := handler.NewUserHandler(userSvc)

	r := gin.Default()
	router.SetupRouter(r, &router.Handler{
		AuthHandler:    authH,
		UserHandler:    userH,
		UserRepository: userRepo,
		JWTService:     jwtSvc,
		CORS:           middleware.CORS(cfg.CORS),
		ValidateOrigin: middleware.ValidateOrigin(cfg.CORS),
	})

	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
