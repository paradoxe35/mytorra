package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpPkg "github.com/paradoxe35/mytorra/packages/http"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/http/handlers"
	httpMiddleware "github.com/paradoxe35/mytorra/services/core/internal/adapters/http/middleware"
)

type Router struct {
	authHandler         *handlers.AuthHandler
	userHandler         *handlers.UserHandler
	areaHandler         *handlers.AreaHandler
	jwksHandler         *handlers.JWKSHandler
	oauthHandler        *handlers.OAuthHandler
	subscriptionHandler *handlers.SubscriptionHandler
	billingHandler      *handlers.BillingHandler
	authMiddleware      *httpMiddleware.AuthMiddleware
	translator          *i18n.Translator
	defaultLanguage     string
}

func NewRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	areaHandler *handlers.AreaHandler,
	jwksHandler *handlers.JWKSHandler,
	oauthHandler *handlers.OAuthHandler,
	subscriptionHandler *handlers.SubscriptionHandler,
	billingHandler *handlers.BillingHandler,
	authMiddleware *httpMiddleware.AuthMiddleware,
	translator *i18n.Translator,
	defaultLanguage string,
) *Router {
	return &Router{
		authHandler:         authHandler,
		userHandler:         userHandler,
		areaHandler:         areaHandler,
		jwksHandler:         jwksHandler,
		oauthHandler:        oauthHandler,
		subscriptionHandler: subscriptionHandler,
		billingHandler:      billingHandler,
		authMiddleware:      authMiddleware,
		translator:          translator,
		defaultLanguage:     defaultLanguage,
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
	r.Use(httpPkg.Language(rt.translator, rt.defaultLanguage)) // Add language middleware

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

			// OAuth routes
			r.Post("/oauth/init", rt.oauthHandler.InitiateOAuth)
			r.Post("/oauth/callback", rt.oauthHandler.HandleCallback)
		})

		// Webhook routes (public but verified by signature)
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/stripe", rt.billingHandler.HandleStripeWebhook)
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

			// OAuth routes that require authentication
			r.Route("/auth/oauth", func(r chi.Router) {
				r.Get("/profiles", rt.oauthHandler.GetOAuthProfiles)
				r.Post("/link", rt.oauthHandler.LinkOAuthAccount)
				r.Delete("/unlink", rt.oauthHandler.UnlinkOAuthAccount)
			})

			// Subscription routes
			r.Route("/subscriptions", func(r chi.Router) {
				r.Get("/", rt.subscriptionHandler.GetUserSubscriptions)
				r.Get("/{id}", rt.subscriptionHandler.GetSubscription)
				r.Post("/", rt.subscriptionHandler.CreateSubscription)
				r.Put("/{id}/upgrade", rt.subscriptionHandler.UpgradeSubscription)
				r.Put("/{id}/downgrade", rt.subscriptionHandler.DowngradeSubscription)
				r.Put("/{id}/pause", rt.subscriptionHandler.PauseSubscription)
				r.Put("/{id}/resume", rt.subscriptionHandler.ResumeSubscription)
				r.Delete("/{id}", rt.subscriptionHandler.CancelSubscription)
				r.Get("/{id}/usage", rt.subscriptionHandler.GetUsage)
				r.Post("/{id}/usage", rt.subscriptionHandler.UpdateUsage)
			})

			// Subscription tier routes
			r.Route("/subscription-tiers", func(r chi.Router) {
				r.Get("/", rt.subscriptionHandler.GetAvailableTiers)
				r.Get("/{code}", rt.subscriptionHandler.GetTierDetails)
			})

			// Billing routes
			r.Route("/billing", func(r chi.Router) {
				r.Get("/cycles", rt.billingHandler.GetBillingCycles)
				r.Get("/cycles/{id}", rt.billingHandler.GetBillingCycle)
				r.Get("/invoices", rt.billingHandler.GetInvoices)
				r.Get("/invoices/{id}", rt.billingHandler.GetInvoice)
				r.Get("/invoices/{id}/download", rt.billingHandler.DownloadInvoice)
				r.Post("/payment-methods", rt.billingHandler.AddPaymentMethod)
				r.Get("/payment-methods", rt.billingHandler.GetPaymentMethods)
				r.Delete("/payment-methods/{id}", rt.billingHandler.DeletePaymentMethod)
				r.Put("/payment-methods/{id}/default", rt.billingHandler.SetDefaultPaymentMethod)
			})
		})
	})

	return r
}
