package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/services/core/internal/adapters/http/handlers"
	httpMiddleware "github.com/paradoxe35/torra/services/core/internal/adapters/http/middleware"
)

type Router struct {
	authHandler    *handlers.AuthHandler
	userHandler    *handlers.UserHandler
	areaHandler    *handlers.AreaHandler
	jwksHandler    *handlers.JWKSHandler
	authMiddleware *httpMiddleware.AuthMiddleware
}

func NewRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	areaHandler *handlers.AreaHandler,
	jwksHandler *handlers.JWKSHandler,
	authMiddleware *httpMiddleware.AuthMiddleware,
) *Router {
	return &Router{
		authHandler:    authHandler,
		userHandler:    userHandler,
		areaHandler:    areaHandler,
		jwksHandler:    jwksHandler,
		authMiddleware: authMiddleware,
	}
}

func (rt *Router) Setup() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(httpPkg.SecurityHeaders)
	r.Use(httpPkg.CORS)

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

	// JWKS endpoints (public - for other services to verify JWTs)
	r.Route("/.well-known", func(r chi.Router) {
		r.Get("/jwks.json", rt.jwksHandler.GetJWKS)       // JSON Web Key Set
		r.Get("/public-key", rt.jwksHandler.GetPublicKey) // Current active public key in PEM format
		r.Get("/keys/info", rt.jwksHandler.GetKeyInfo)    // Key information (without actual keys)
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Authentication routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", rt.authHandler.Register)
			r.Post("/login", rt.authHandler.Login)
			r.Post("/refresh", rt.authHandler.RefreshToken)
			r.Post("/forgot-password", rt.authHandler.ForgotPassword)
			r.Post("/reset-password", rt.authHandler.ResetPassword)
			r.Get("/verify-email", rt.authHandler.VerifyEmail)
			r.Post("/verify-phone", rt.authHandler.VerifyPhone)
		})

		// Area routes (public)
		r.Route("/areas", func(r chi.Router) {
			r.Post("/detect", rt.areaHandler.DetectArea) // Area detection endpoint
			r.Get("/", rt.areaHandler.GetActiveAreas)    // List active areas
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// Add authentication middleware
			r.Use(rt.authMiddleware.Authenticate)

			// Auth routes that require authentication
			r.Post("/auth/logout", rt.authHandler.Logout)

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", rt.userHandler.GetCurrentUser)
				r.Put("/me", rt.userHandler.UpdateCurrentUser)
				r.Delete("/me", rt.userHandler.DeleteCurrentUser)
				r.Post("/me/mfa/enable", rt.userHandler.EnableMFA)
				r.Post("/me/mfa/disable", rt.userHandler.DisableMFA)
				r.Put("/me/area", rt.userHandler.UpdateArea)
				r.Get("/me/verification-status", rt.userHandler.GetVerificationStatus)
				r.Post("/me/verify-id", rt.userHandler.VerifyGovernmentID)
			})

			// TODO: Add more protected routes for payments, subscriptions, etc.
		})
	})

	return r
}
