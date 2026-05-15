package src

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"viralforge/backend/src/adapter"
)

const timeout = 30 * time.Second

// NewServer creates and configures the HTTP server.
func NewServer(cfg *Config, store *adapter.PostgresStore, email adapter.EmailService) *Server {
	return &Server{
		cfg:    cfg,
		store:  store,
		email:  email,
		router: chi.NewMux(),
	}
}

// Server holds the HTTP server state.
type Server struct {
	cfg    *Config
	store  *adapter.PostgresStore
	email  adapter.EmailService
	router chi.Router
}

// Setup configures routing and middleware.
func (s *Server) Setup() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(timeout))

	s.router.Get("/health", s.healthCheck)
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
