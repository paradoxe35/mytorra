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

	locationService := services.NewLocationService(
		areaService,
		appLogger,
	)

	userService := services.NewUserService(
		userRepo,
		areaService,
		locationService,
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
	areaHandler := handlers.NewAreaHandler(areaService, locationService, appLogger)
	jwksHandler := handlers.NewJWKSHandler(jwtKeyRepo, appLogger, cacheClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtKeyRepo, appLogger)

	// Setup router
	router := adaptersHTTP.NewRouter(authHandler, userHandler, areaHandler, jwksHandler, authMiddleware)
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

	// Start key rotation scheduler
	stopKeyRotation := startKeyRotationScheduler(keyManagementService, appLogger)
	defer stopKeyRotation()

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

// startKeyRotationScheduler starts a background goroutine that handles automatic key rotation
// and cleanup of expired keys. It runs daily and returns a stop function.
func startKeyRotationScheduler(keyService *services.KeyManagementService, logger logger.Interface) func() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	done := make(chan bool)

	// Run immediately on startup
	go func() {
		ctx := context.Background()
		performKeyMaintenance(ctx, keyService, logger)
	}()

	// Then run periodically
	go func() {
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				performKeyMaintenance(ctx, keyService, logger)
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Return stop function
	return func() {
		close(done)
	}
}

func performKeyMaintenance(ctx context.Context, keyService *services.KeyManagementService, logger logger.Interface) {
	logger.Info("Starting key maintenance tasks")

	// Check if we need to rotate keys
	activeKey, err := keyService.GetActiveKey(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get active key")
		// Try to generate a new key if no active key exists
		if err := keyService.GenerateNewKeyPair(ctx); err != nil {
			logger.WithError(err).Error("Failed to generate new key pair")
		}
		return
	}

	// Check if key expires within 5 days
	daysUntilExpiry := time.Until(activeKey.ExpiresAt).Hours() / 24
	logger.Info("Active key status",
		"key_id", activeKey.KeyID,
		"expires_in_days", int(daysUntilExpiry))

	if daysUntilExpiry <= 5 {
		logger.Info("Active key expiring soon, rotating keys")
		if err := keyService.RotateKeys(ctx); err != nil {
			logger.WithError(err).Error("Failed to rotate keys")
		} else {
			logger.Info("Keys rotated successfully")
		}
	}

	// Cleanup expired keys
	if err := keyService.CleanupExpiredKeys(ctx); err != nil {
		logger.WithError(err).Error("Failed to cleanup expired keys")
	} else {
		logger.Info("Expired keys cleanup completed")
	}
}
