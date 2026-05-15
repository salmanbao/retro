package src

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"viralforge/backend/src/adapter"
	"viralforge/backend/src/handler"
	authMiddleware "viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

const timeout = 30 * time.Second

// NewServer creates and configures the HTTP server.
func NewServer(cfg *Config, db *adapter.PostgresStore, email adapter.EmailService) *Server {
	return &Server{
		cfg:   cfg,
		store: db,
		email: email,
		router: chi.NewMux(),
	}
}

// Server holds the HTTP server state.
type Server struct {
	cfg    *Config
	store  *adapter.PostgresStore
	email  adapter.EmailService
	router *chi.Mux
}

// Setup configures routing and middleware.
func (s *Server) Setup() {
	s.router.Use(chiMiddleware.RequestID)
	s.router.Use(chiMiddleware.RealIP)
	s.router.Use(chiMiddleware.Logger)
	s.router.Use(chiMiddleware.Recoverer)
	s.router.Use(chiMiddleware.Timeout(timeout))

	// Auto-migrate domain models
	_ = s.store.AutoMigrate()

	s.router.Get("/health", s.healthCheck)

	// Create services
	authSvc := service.NewAuthService(
		s.store.UserRepository(),
		s.store.SessionRepository(),
		s.store.TokenRepository(),
		s.email,
		s.cfg.BaseURL,
	)
	loginHistorySvc := service.NewLoginHistoryService(s.store)

	var encryptionKey []byte
	if s.cfg.EncryptionKey != "" {
		encryptionKey = []byte(s.cfg.EncryptionKey)
	} else {
		// Use a default key for development (NOT for production)
		encryptionKey = []byte("default-32-byte-key-for-dev!")
	}
	twoFactorSvc := service.NewTwoFactorService(s.store, encryptionKey)

	profileSvc := service.NewProfileService(
		s.store.ProfileRepository(),
		s.store.UserRepository(),
	)

	sessionSvc := service.NewSessionService(
		s.store.SessionRepository(),
		s.store.ProfileRepository(),
	)

	// Create handlers
	authHandler := handler.NewAuthHandler(authSvc, loginHistorySvc, twoFactorSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	profileHandler := handler.NewProfileHandler(profileSvc)
	twoFactorHandler := handler.NewTwoFactorHandler(twoFactorSvc, authSvc)

	// Create middleware
	authMw := authMiddleware.NewAuthMiddleware(s.store.SessionRepository(), s.store.UserRepository())

	s.router.Route("/api/v1/auth", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})
	// Protected auth routes
	s.router.Route("/api/v1/auth/me", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/me", authHandler.Me)
	})
	// Protected profile routes
	s.router.Route("/api/v1/profiles", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		profileHandler.RegisterRoutes(r)
	})
	// Protected session routes
	s.router.Route("/api/v1/sessions", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Patch("/active", sessionHandler.SwitchActiveProfile)
		sessionHandler.RegisterRoutes(r)
	})
	// 2FA routes
	s.router.Route("/api/v1/auth/2fa", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		twoFactorHandler.RegisterRoutes(r)
	})
	// Login history route
	s.router.Route("/api/v1/auth/login-history", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		authHandler.RegisterLoginHistoryRoutes(r)
	})
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Run starts the HTTP server.
func (s *Server) Run(ctx context.Context, addr string) error {
	s.Setup()
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	go srv.ListenAndServe()
	<-ctx.Done()
	return srv.Shutdown(ctx)
}