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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"

	httpPkg "github.com/paradoxe35/torra/packages/http"
)

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(cfg.Server.ReadTimeout))

	// Health check endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Check database, redis, etc.
		httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
			"status": "ready",
		})
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handleRegister)
			r.Post("/login", handleLogin)
			r.Post("/refresh", handleRefresh)
			r.Post("/logout", handleLogout)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// Add authentication middleware here
			// r.Use(authMiddleware)

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", handleGetCurrentUser)
				r.Put("/me", handleUpdateCurrentUser)
				r.Delete("/me", handleDeleteCurrentUser)
			})
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Core Service on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

type Config struct {
	Server ServerConfig
}

type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../../configs")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("MYTORRA")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Placeholder handlers - to be implemented
func handleRegister(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Registration endpoint not yet implemented")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Login endpoint not yet implemented")
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Token refresh endpoint not yet implemented")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Logout endpoint not yet implemented")
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Get current user endpoint not yet implemented")
}

func handleUpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Update user endpoint not yet implemented")
}

func handleDeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Delete user endpoint not yet implemented")
}
