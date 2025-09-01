package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Shared packages
	"github.com/paradoxe35/torra/packages/cache"
	"github.com/paradoxe35/torra/packages/config"
	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/events"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"

	// Internal packages
	adaptersHTTP "github.com/paradoxe35/torra/services/core/internal/adapters/http"
	"github.com/paradoxe35/torra/services/core/internal/adapters/http/handlers"
	"github.com/paradoxe35/torra/services/core/internal/adapters/http/middleware"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/repositories"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Service.Name, cfg.Service.Environment)

	// Initialize translator
	translator, err := i18n.New("en")
	if err != nil {
		appLogger.Fatal("Failed to initialize translator")
	}

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Initialize cache
	cacheClient, err := cache.NewRedisCacheAdapter(&cfg.Redis, "core")
	if err != nil {
		appLogger.Fatal("Failed to connect to cache")
	}
	defer func() {
		if closer, ok := cacheClient.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	areaRepo := repositories.NewAreaRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	jwtKeyRepo := repositories.NewJWTKeyRepository(db)

	// Initialize event publisher
	eventPublisher, err := events.NewRabbitMQEventPublisher(&cfg.RabbitMQ, appLogger)
	if err != nil {
		appLogger.Warn(fmt.Sprintf("Failed to connect to RabbitMQ: %v. Using simple publisher", err))
		eventPublisher = &events.SimpleEventPublisher{Logger: appLogger}
	}
	defer eventPublisher.Close()

	// Initialize application services
	areaService := services.NewAreaService(
		areaRepo,
		eventPublisher,
		cacheClient,
		appLogger,
	)

	userService := services.NewUserService(
		userRepo,
		areaService,
		eventPublisher,
		cacheClient,
		appLogger,
		translator,
	)

	// Initialize key management service first to ensure keys exist
	keyManagementService := services.NewKeyManagementService(jwtKeyRepo, appLogger)

	// Ensure at least one active JWT key exists
	if err := keyManagementService.InitializeKeys(context.Background()); err != nil {
		appLogger.Fatal("Failed to initialize JWT keys", "error", err)
	}

	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		jwtKeyRepo,
		cacheClient,
		appLogger,
		translator,
		cfg.JWT,
	)

	// Initialize HTTP handlers
	authHandler := handlers.NewAuthHandler(authService, userService, appLogger)
	userHandler := handlers.NewUserHandler(userService, appLogger)
	jwksHandler := handlers.NewJWKSHandler(jwtKeyRepo, appLogger, cacheClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtKeyRepo, appLogger)

	// Setup router
	router := adaptersHTTP.NewRouter(authHandler, userHandler, jwksHandler, authMiddleware)
	r := router.Setup()

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		appLogger.Info(fmt.Sprintf("Starting %s on port %d", cfg.Service.Name, cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownTimeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown")
	}

	appLogger.Info("Server exited")
}
