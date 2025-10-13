package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"sage-of-elements-backend/internal/adapters/cache/redis"
	"sage-of-elements-backend/internal/adapters/primary/http/middleware"
	"sage-of-elements-backend/internal/adapters/storage/postgres"
	"sage-of-elements-backend/internal/modules/character"
	"sage-of-elements-backend/internal/modules/combat"
	"sage-of-elements-backend/internal/modules/deck"
	"sage-of-elements-backend/internal/modules/fusion"
	"sage-of-elements-backend/internal/modules/game_data"
	"sage-of-elements-backend/internal/modules/player"
	"sage-of-elements-backend/internal/modules/pve"
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/appconfig"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
	"sage-of-elements-backend/pkg/appresponse"
	"sage-of-elements-backend/pkg/appvalidator"
	"sage-of-elements-backend/pkg/platform/apppostgres"
	"sage-of-elements-backend/pkg/platform/appredis"
)

func main() {
	// --- 1. Setup ---
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	var appLogger applogger.Logger
	if cfg.Server.Mode == "development" {
		appLogger = applogger.NewPrettyLogger()
	} else {
		appLogger = applogger.NewSlogLogger()
	}
	appLogger.Info("Starting Sage of the Elements Backend...", "env", cfg.Server.Mode)

	redisClient, err := appredis.NewConnection(cfg.Redis, appLogger)
	if err != nil {
		appLogger.Error("could not connect to Redis", err)
		os.Exit(1)
	}

	// --- 2. Platforms Connection ---
	db, err := apppostgres.NewConnection(cfg.Postgres.Primary, appLogger)
	if err != nil {
		appLogger.Error("could not connect to database", err)
		os.Exit(1)
	}
	appLogger.Success("Database connection successful.")

	if err = postgres.AutoMigrate(db, appLogger); err != nil {
		appLogger.Error("could not auto migrate database:", err)
		os.Exit(1)
	}
	if err = postgres.Seed(db); err != nil {
		appLogger.Error("could not seed database:", err)
		os.Exit(1)
	}

	// --- 3. Dependency Injection ---
	appValidator := appvalidator.New()
	authSvc := appauth.NewAuthService(cfg.Auth.JWTAccessSecret, cfg.Auth.JWTRefreshSecret)

	playerRepo := postgres.NewPlayerRepository(db)
	playerSvc := player.NewPlayerService(appLogger, authSvc, playerRepo)
	playerHandler := player.NewPlayerHandler(appValidator, playerSvc)

	gameDataDbRepo := postgres.NewGameDataRepository(db)
	gameDataCacheRepo := redis.NewGameDataCacheRepository(redisClient)
	gameDataSvc := game_data.NewGameDataService(appLogger, gameDataDbRepo, gameDataCacheRepo)
	gameDataHandler := game_data.NewGameDataHandler(gameDataSvc)

	characterRepo := postgres.NewCharacterRepository(db)
	characterSvc := character.NewCharacterService(appLogger, characterRepo, gameDataDbRepo)
	characterHandler := character.NewCharacterHandler(appValidator, characterSvc)

	deckRepo := postgres.NewDeckRepository(db)
	deckSvc := deck.NewDeckService(appLogger, deckRepo, characterRepo)
	deckHandler := deck.NewDeckHandler(appValidator, deckSvc)

	fusionRepo := postgres.NewFusionRepository(db)
	fusionSvc := fusion.NewFusionService(appLogger, db, fusionRepo, characterRepo, gameDataDbRepo)
	fusionHandler := fusion.NewFusionHandler(appValidator, fusionSvc)

	pveRepo := postgres.NewPveRepository(db)
	pveSvc := pve.NewPveService(pveRepo)
	pveHandler := pve.NewPveHandler(pveSvc)

	enemyRepo := postgres.NewEnemyRepository(db)
	combatRepo := postgres.NewCombatRepository(db)
	combatSvc := combat.NewCombatService(appLogger, combatRepo, characterRepo, enemyRepo, pveRepo)
	combatHandler := combat.NewCombatHandler(appValidator, combatSvc)

	// --- 4. Setup Fiber App & Routes ---
	app := fiber.New(fiber.Config{
		AppName: "Sage of the Elements API " + cfg.App.Version,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*apperrors.AppError); ok {
				return appresponse.Error(c, appErr)
			}
			systemErr := apperrors.SystemErrorWithDetails("An unexpected system error occurred", err.Error())
			appLogger.Error("Unhandled error", systemErr)
			return appresponse.Error(c, systemErr)
		},
	})

	// --- สร้าง Middleware Instances ---
	corsMiddleware := middleware.CORSMiddleware()
	logMiddleware := middleware.LoggerMiddleware(appLogger)
	authMiddleware := middleware.AuthMiddleware(authSvc)

	// --- 5. ติดตั้ง Middlewares ---
	app.Use(corsMiddleware)             // จัดการเรื่อง CORS
	app.Use(logMiddleware)              // จัดการเรื่อง Log
	app.Use(limiter.New(limiter.Config{ // ป้องกันการยิง Request ถี่ๆ
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	apiV1 := app.Group("/api/v1")
	apiV1.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"status": "ok"})
	})

	// --- สร้าง Group หลักสำหรับแต่ละ Module ---
	playerGroup := apiV1.Group("/players")
	characterGroup := apiV1.Group("/characters")
	deckGroup := apiV1.Group("/decks")
	gameDataGroup := apiV1.Group("/game-data")
	fusionGroup := apiV1.Group("/fusion")
	pveGroup := apiV1.Group("/pve")
	matchGroup := apiV1.Group("/matches")
	// --- Public Routes ---
	playerHandler.RegisterPublicRoutes(playerGroup)

	playerGroup.Use(authMiddleware)
	characterGroup.Use(authMiddleware)
	deckGroup.Use(authMiddleware)
	gameDataGroup.Use(authMiddleware)
	fusionGroup.Use(authMiddleware)
	pveGroup.Use(authMiddleware)
	matchGroup.Use(authMiddleware)

	// --- Protected Routes ---

	// ลงทะเบียน Protected Routes
	playerHandler.RegisterProtectedRoutes(playerGroup)
	characterHandler.RegisterProtectedRoutes(characterGroup)
	deckHandler.RegisterProtectedRoutes(deckGroup)
	gameDataHandler.RegisterProtectedRoutes(gameDataGroup)

	fusionHandler.RegisterProtectedRoutes(fusionGroup)
	pveHandler.RegisterProtectedRoutes(pveGroup)
	combatHandler.RegisterProtectedRoutes(matchGroup)

	// --- 5. Start Server & Graceful Shutdown ---
	go func() {
		listenAddr := ":" + cfg.Server.AppPort
		appLogger.Info("Server starting...", "port", cfg.Server.AppPort)
		if err := app.Listen(listenAddr); err != nil {
			appLogger.Error("Server failed to start", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	appLogger.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		appLogger.Error("Server shutdown failed", err)
	}

	appLogger.Info("Server gracefully stopped")
}
