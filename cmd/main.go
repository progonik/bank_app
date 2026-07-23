package main

import (
	"context"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/prodonik/bank_app/config"
	_ "github.com/prodonik/bank_app/docs"
	appcity "github.com/prodonik/bank_app/internal/application/city"
	appent "github.com/prodonik/bank_app/internal/application/entrepreneur"
	appifut "github.com/prodonik/bank_app/internal/application/ifut_code"
	appinn "github.com/prodonik/bank_app/internal/application/inn"
	appuser "github.com/prodonik/bank_app/internal/application/user"
	"github.com/prodonik/bank_app/internal/infrastructure/auth"
	"github.com/prodonik/bank_app/internal/infrastructure/birdarcha"
	"github.com/prodonik/bank_app/internal/infrastructure/bitrix"
	"github.com/prodonik/bank_app/internal/infrastructure/database"
	"github.com/prodonik/bank_app/internal/infrastructure/repository"
	"github.com/prodonik/bank_app/internal/infrastructure/sqb"
	"github.com/prodonik/bank_app/internal/interfaces/api"
	"github.com/prodonik/bank_app/internal/interfaces/api/handler"
)

// @title Bank App API
// @version 1.0
// @description Authentication service for the Bank App
// @host bank-app-back.compile-me.uz
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Run migrations
	m, err := migrate.New("file://internal/infrastructure/database/migrations", cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Database connection
	db, err := database.NewConnection(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Infrastructure
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	cityRepo := repository.NewCityRepository(db)
	innRepo := repository.NewInnRepository(db)
	ifutCodeRepo := repository.NewIfutCodeRepository(db)
	entrepreneurRepo := repository.NewEntrepreneurRepository(db)

	// SQB client
	sqbClient := sqb.NewClient(cfg.SQBBaseURL, cfg.SQBLocalAddr)
	bitrixClient := bitrix.NewClient(cfg.BitrixWebhookURL)

	// Application
	userService := appuser.NewService(userRepo, sessionRepo, jwtService)
	cityService := appcity.NewService(cityRepo)
	innService := appinn.NewService(innRepo)
	ifutCodeService := appifut.NewService(ifutCodeRepo)
	entrepreneurService := appent.NewService(entrepreneurRepo, innRepo, ifutCodeRepo, sqbClient, bitrixClient, db)

	// Birdarcha syncers (token is loaded from DB at each sync cycle)
	birdarchaClient := birdarcha.NewClient(cfg.BirdarchaBaseURL, "")
	syncer := birdarcha.NewSyncer(birdarchaClient, entrepreneurService, db, cfg.BirdarchaSyncInterval, cfg.BirdarchaCutoffDate)
	individualBirdarchaClient := birdarcha.NewClient(cfg.BirdarchaBaseURL, "")
	individualSyncer := birdarcha.NewIndividualSyncer(individualBirdarchaClient, entrepreneurService, db, cfg.BirdarchaSyncInterval, cfg.BirdarchaCutoffDate)
	syncCtx, syncCancel := context.WithCancel(context.Background())
	defer syncCancel()
	go syncer.Start(syncCtx)
	go individualSyncer.Start(syncCtx)

	// Interfaces
	userHandler := handler.NewUserHandler(userService)
	cityHandler := handler.NewCityHandler(cityService)
	innHandler := handler.NewInnHandler(innService)
	ifutCodeHandler := handler.NewIfutCodeHandler(ifutCodeService)
	entrepreneurHandler := handler.NewEntrepreneurHandler(entrepreneurService)
	router := api.NewRouter(userHandler, cityHandler, innHandler, ifutCodeHandler, entrepreneurHandler, jwtService)

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
