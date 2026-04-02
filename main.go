package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/auth"
	"github.com/spendly/backend/internal/bootstrap"
	"github.com/spendly/backend/internal/config"
	handler "github.com/spendly/backend/internal/handler/http"
	appMiddleware "github.com/spendly/backend/internal/middleware"
	"github.com/spendly/backend/internal/repository/postgres"
	"github.com/spendly/backend/internal/scheduler"
	"github.com/spendly/backend/internal/service"
)

func main() {
	// 1. Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Bootstrap Database
	db, err := bootstrap.NewPostgresDB(ctx, bootstrap.ConfigDB{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute,
		ConnMaxIdleTime: time.Duration(cfg.Database.ConnMaxIdleTime) * time.Minute,
	})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	// 3. Initialize Repositories
	userRepo := postgres.NewUserRepository(db.DB)
	txnRepo := postgres.NewTransactionRepository(db.DB)
	snapRepo := postgres.NewAnalysisRepository(db.DB)
	insightRepo := postgres.NewInsightRepository(db.DB)
	budgetRepo := postgres.NewBudgetRepository(db.DB)

	// 4. Initialize Auth & AI Clients
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24*time.Hour)
	_ = auth.NewGoogleAuthenticator(cfg.GoogleClientID)

	gemini, err := ai.NewGeminiClient(ctx, cfg.GeminiApiKey)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Gemini client: %v", err)
	}
	defer gemini.Close()

	analysisPipeline := service.NewAnalysisPipeline(gemini, txnRepo, snapRepo, insightRepo)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler()
	txnHandler := handler.NewTransactionHandler(txnRepo)
	budgetHandler := handler.NewBudgetHandler(budgetRepo)
	reportHandler := handler.NewReportHandler(snapRepo)
	insightHandler := handler.NewInsightHandler(insightRepo)

	// 6. Setup Router & Middlewares
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Rate Limiting (10 requests per second per IP)
	limiter := appMiddleware.NewIPRateLimiter(10, 20)
	r.Use(appMiddleware.RateLimitMiddleware(limiter))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public Routes
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "pong"}`))
	})
	r.Post("/auth/google", authHandler.GoogleLogin)

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.AuthMiddleware(jwtManager))
		
		r.Get("/transactions", txnHandler.ListTransactions)
		r.Get("/transactions/{id}", txnHandler.GetByID)
		
		r.Get("/budgets", budgetHandler.ListActiveBudgets)
		r.Post("/budgets", budgetHandler.CreateBudget)
		
		r.Get("/reports/latest", reportHandler.GetLatestReport)
		
		r.Get("/insights/unread", insightHandler.ListUnread)
		r.Patch("/insights/{id}/read", insightHandler.MarkAsRead)
	})

	// 7. Initialize & Start Scheduler
	sched := scheduler.NewScheduler(snapRepo)
	sched.AddMonthlyAnalysisJob(analysisPipeline, userRepo)
	sched.AddDailyBudgetJob(userRepo, budgetRepo)
	sched.Start()
	defer sched.Stop()

	// 8. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
	}

	go func() {
		log.Printf("🚀 Server is running on port %s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Listen error: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("👋 Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✨ Server stopped gracefully")
}
