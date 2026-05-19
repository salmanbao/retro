package src

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"viralforge/backend/src/adapter"
	"viralforge/backend/src/handler"
	assetHandler "viralforge/backend/src/handler/asset"
	campaignHandler "viralforge/backend/src/handler/campaign"
	creativeBriefHandler "viralforge/backend/src/handler/creative_brief"
	onboardingHandler "viralforge/backend/src/handler/onboarding"
	authMiddleware "viralforge/backend/src/middleware"
	onboardingRepo "viralforge/backend/src/repository/onboarding"
	"viralforge/backend/src/service"
	"viralforge/backend/src/service/onboarding"
)

const timeout = 30 * time.Second

// NewServer creates and configures the HTTP server.
func NewServer(cfg *Config, db *adapter.PostgresStore, email adapter.EmailService) *Server {
	return &Server{
		cfg:    cfg,
		store:  db,
		email:  email,
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

	enrichmentSvc := service.NewProfileEnrichmentService(
		s.store.ProfileEnrichmentRepository(),
		s.store.ProfileRepository(),
	)

	portfolioSvc := service.NewPortfolioService(
		s.store.PortfolioItemRepository(),
		s.store.ProfileRepository(),
	)

	audienceSvc := service.NewAudienceService(
		s.store.AudienceDataRepository(),
		s.store.ProfileRepository(),
	)

	verificationSvc := service.NewVerificationService(
		s.store.FollowerVerificationRepository(),
		s.store.ProfileRepository(),
	)

	kycSvc := service.NewKYCService(
		s.store.KYCStatusRepository(),
		s.store.ProfileRepository(),
	)

	payoutSvc := service.NewPayoutService(
		s.store.PayoutPreferencesRepository(),
		s.store.ProfileRepository(),
	)

	sessionSvc := service.NewSessionService(
		s.store.SessionRepository(),
		s.store.ProfileRepository(),
	)

	campaignSvc := service.NewCampaignService(
		s.store.CampaignRepository(),
		s.store.CampaignAssetRepository(),
	)

	// Creative Brief and Asset services
	briefSvc := service.NewCreativeBriefService(
		s.store.CreativeBriefRepository(),
		s.store.CampaignRepository(),
	)
	assetSvc := service.NewAssetService(
		s.store.AssetRepository(),
		s.store.CampaignRepository(),
	)

	onboardingSvc := onboarding.NewService(
		onboardingRepo.NewTemplateRepo(s.store.DB()),
		onboardingRepo.NewProgressRepo(s.store.DB()),
		onboardingRepo.NewStepRepo(s.store.DB()),
	)
	activationSvc := onboarding.NewActivationService(
		onboardingRepo.NewTemplateRepo(s.store.DB()),
		onboardingRepo.NewProgressRepo(s.store.DB()),
		onboardingRepo.NewStepRepo(s.store.DB()),
	)

	// Create handlers
	authHandler := handler.NewAuthHandler(authSvc, loginHistorySvc, twoFactorSvc)
	campaignHdlr := campaignHandler.NewHandler(campaignSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	profileHandler := handler.NewProfileHandler(profileSvc)
	profileEnrichmentHandler := handler.NewProfileEnrichmentHandler(enrichmentSvc)
	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)
	audienceHandler := handler.NewAudienceHandler(audienceSvc)
	verificationHandler := handler.NewVerificationHandler(verificationSvc)
	kycHandler := handler.NewKYCHandler(kycSvc)
	payoutHandler := handler.NewPayoutHandler(payoutSvc)
	twoFactorHandler := handler.NewTwoFactorHandler(twoFactorSvc, authSvc)
	onboardingHandler := onboardingHandler.NewHandler(onboardingSvc, activationSvc)
	briefHdlr := creativeBriefHandler.NewHandler(briefSvc, campaignSvc, s.store.ProfileRepository())
	assetHdlr := assetHandler.NewHandler(assetSvc, campaignSvc, s.store.ProfileRepository())

	// Create middleware
	authMw := authMiddleware.NewAuthMiddleware(s.store.SessionRepository(), s.store.UserRepository())
	ownershipMw := authMiddleware.NewOwnershipMiddleware(s.store.ProfileRepository())
	adminMw := authMiddleware.NewAdminMiddleware(s.store.UserRepository())
	profileTypeMw := authMiddleware.NewProfileTypeMiddleware(s.store.ProfileRepository())

	s.router.Route("/api/v1/auth", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})
	// Protected /me route - must be explicitly protected since RegisterRoutes handles it without middleware
	s.router.Route("/api/v1/auth/me", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/", authHandler.Me)
	})
	// Protected profile routes
	s.router.Route("/api/v1/profiles", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		profileHandler.RegisterRoutes(r)
		r.Route("/{id}/details", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			r.Get("/", profileEnrichmentHandler.GetDetails)
			r.Patch("/", profileEnrichmentHandler.UpdateDetails)
		})
		r.Route("/{id}/portfolio", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			r.Use(profileTypeMw.RequireEditor) // Editor profile required
			r.Get("/", portfolioHandler.GetPortfolio)
			r.Post("/", portfolioHandler.CreatePortfolioItem)
			r.Patch("/{itemId}", portfolioHandler.UpdatePortfolioItem)
			r.Delete("/{itemId}", portfolioHandler.DeletePortfolioItem)
		})
		r.Route("/{id}/audience", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			r.Use(profileTypeMw.RequireInfluencer) // Influencer profile required
			r.Get("/", audienceHandler.GetAudience)
			r.Put("/", audienceHandler.UpdateAudience)
		})
		r.Route("/{id}/verification", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			r.Use(profileTypeMw.RequireInfluencer) // Influencer profile required
			r.Get("/", verificationHandler.GetVerification)
			r.Post("/", verificationHandler.SubmitVerification)
		})
		r.Route("/{id}/kyc", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			r.Get("/", kycHandler.GetKYCStatus)
		})
		r.Route("/{id}/payout", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			r.Get("/", payoutHandler.GetPayoutPreferences)
			r.Put("/", payoutHandler.UpdatePayoutPreferences)
		})
		r.Route("/{id}/onboarding", func(r chi.Router) {
			r.Use(ownershipMw.RequireOwnership)
			onboardingHandler.RegisterPublicRoutes(r)
		})
	})
	// Admin routes
	s.router.Route("/api/v1/admin/profiles", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Use(adminMw.RequireAdmin) // Admin access required
		r.Put("/{profileId}/verification/review", verificationHandler.AdminReviewVerification)
		r.Put("/{profileId}/kyc", kycHandler.AdminUpdateKYC)
		r.Post("/{profileId}/onboarding/activate", onboardingHandler.AdminActivate)
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
	// Campaign routes
	s.router.Route("/api/v1/campaigns", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		campaignHdlr.RegisterRoutes(r)
	})
	// Creative brief routes
	s.router.Route("/api/v1/campaigns", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		briefHdlr.RegisterRoutes(r)
	})
	// Asset routes
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		assetHdlr.RegisterRoutes(r)
	})
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Run starts the HTTP server.
func (s *Server) Run(ctx context.Context, addr string) error {
	s.Setup()

	// Start campaign lifecycle worker for automatic published->active transitions
	go s.startCampaignLifecycleWorker(ctx)

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

// startCampaignLifecycleWorker periodically transitions published campaigns to active.
func (s *Server) startCampaignLifecycleWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			campaignSvc := service.NewCampaignService(
				s.store.CampaignRepository(),
				s.store.CampaignAssetRepository(),
			)
			count, err := campaignSvc.TransitionPublishedToActive(ctx)
			if err != nil {
				log.Printf("Campaign lifecycle worker error: %v", err)
			} else if count > 0 {
				log.Printf("Campaign lifecycle worker: transitioned %d campaigns from published to active", count)
			}
		}
	}
}
