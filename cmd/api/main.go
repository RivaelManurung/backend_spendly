package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/ai"
	"github.com/spendly/backend/internal/bootstrap"
	"github.com/spendly/backend/internal/config"
	handlers "github.com/spendly/backend/internal/handler/http"
	"github.com/spendly/backend/internal/repository/postgres"
	"github.com/spendly/backend/internal/scheduler"
	"github.com/spendly/backend/internal/service"

	_ "github.com/spendly/backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Spendly API
// @version 1.0
// @description Backend API for Spendly Personal Finance Tracker.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 1. Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Initialize Database Connection (PostgreSQL via sqlx)
	dbCfg := bootstrap.ConfigDB{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute,
		ConnMaxIdleTime: time.Duration(cfg.Database.ConnMaxIdleTime) * time.Minute,
	}
	db, err := bootstrap.NewPostgresDB(ctx, dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 3. Initialize AI Client (Gemini)
	gemini, err := ai.NewGeminiClient(context.Background(), cfg.GeminiApiKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini: %v", err)
	}
	defer gemini.Close()

	// 4. Initialize Repositories
	repoUser     := postgres.NewUserRepository(db.DB)
	repoSnap     := postgres.NewAnalysisRepository(db.DB)
	repoInsight  := postgres.NewInsightRepository(db.DB)
	repoBudget   := postgres.NewBudgetRepository(db.DB)
	repoTxn      := postgres.NewTransactionRepository(db.DB)
	repoRecurring := postgres.NewRecurringRepository(db.DB)
	repoAccount  := postgres.NewAccountRepository(db.DB)

	// Silence unused variable warnings for repos not yet used in handlers
	_ = repoBudget
	_ = repoAccount

	// 5. Initialize Services
	analysisPipeline := service.NewAnalysisPipeline(
		gemini, repoTxn, repoSnap, repoInsight, repoUser,
	)
	dailyDigestSvc := service.NewDailyDigestService(
		gemini, repoTxn, repoInsight, repoBudget,
	)
	recurringSvc := service.NewRecurringService(
		gemini, repoRecurring, repoTxn, repoInsight,
	)
	netWorthSvc := service.NewNetWorthService(
		gemini, repoInsight, repoUser,
	)

	// 6. Initialize Scheduler & Background Jobs
	sched := scheduler.NewScheduler()
	sched.AddMonthlyAnalysisJob(analysisPipeline, repoUser)
	sched.AddDailyBudgetJob(repoUser, repoBudget)
	sched.AddDailyTasks(repoUser, dailyDigestSvc, recurringSvc, netWorthSvc)
	
	// Start Scheduler
	log.Println("Starting Background Scheduler...")
	sched.Start()
	defer sched.Stop()

	// 7. Initialize Handlers
	analysisHandler := handlers.NewAnalysisHandler(repoSnap)
	insightHandler  := handlers.NewInsightHandler(repoInsight)
	userHandler     := handlers.NewUserHandler(repoUser)

	// 8. Setup Router & Routes
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		// User Profile
		api.GET("/users/:id",  userHandler.GetProfile)
		api.PUT("/users/:id",  userHandler.UpdateProfile)

		// Analysis & Insights
		api.GET("/analysis/latest",        analysisHandler.GetLatestSnapshot)
		api.GET("/insights/latest",         insightHandler.GetLatestInsights)
		api.PATCH("/insights/:id/read",     insightHandler.MarkAsRead)

		// Basic Health Check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "Spendly Backend",
				"version": "1.0.0",
			})
		})
	}

	// Swagger Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 8. Start Server
	log.Printf("Spendly Backend starting on port %s (%s environment)\n", cfg.HTTPPort, cfg.AppEnv)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("Router failed to run: %v", err)
	}
}
