package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/bootstrap"
	"github.com/spendly/backend/internal/config"
	handler "github.com/spendly/backend/internal/handler/http"
	"github.com/spendly/backend/internal/repository"
	"github.com/spendly/backend/internal/scheduler"
	"github.com/spendly/backend/internal/service"
)

func main() {
	// Initialize Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fail loading config: %v", err)
	}

	// Setup SQLite + GORM Database
	db, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Could not start SQLite DB: %v", err)
	}

	catRepo := repository.NewCategoryRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	goalRepo := repository.NewGoalRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)

	ctx := context.Background()

	// === Background Jobs / Cron ===
	cronJob := scheduler.NewDailyJob(db)
	cronJob.StartCron(ctx)

	// === DI Container (Services) ===
	catSvc := service.NewCategoryService(catRepo)
	txSvc := service.NewTransactionService(txRepo)
	aiSvc := service.NewAIService(ctx, txRepo, cfg)
	goalSvc := service.NewGoalService(goalRepo)
	budgetSvc := service.NewBudgetService(budgetRepo, txRepo)
	invoiceSvc := service.NewInvoiceService(invoiceRepo)
	reportSvc := service.NewReportService(txRepo)
	syncSvc := service.NewSyncService(txRepo)

	// === DI Container (Handlers) ===
	catH := handler.NewCategoryHandler(catSvc)
	txH := handler.NewTransactionHandler(txSvc)
	ocrH := handler.NewOCRHandler(aiSvc)
	goalH := handler.NewGoalHandler(goalSvc)
	budgetH := handler.NewBudgetHandler(budgetSvc)
	invoiceH := handler.NewInvoiceHandler(invoiceSvc)
	analyticsH := handler.NewAnalyticsHandler(reportSvc, syncSvc)

	// Create Gin Router
	r := gin.Default()

	// App Routes (All public for local single-user mode)
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong! spendly core running (Local Single-User Mode)",
				"env":     cfg.Env,
			})
		})

		// Categories
		catRoutes := api.Group("/categories")
		{
			catRoutes.POST("/", catH.Create)
			catRoutes.GET("/", catH.List)
		}

		// Transactions
		txRoutes := api.Group("/transactions")
		{
			txRoutes.POST("/", txH.Create)
			txRoutes.GET("/", txH.GetTransactions)
		}

		// AI Analyst & Scanner
		aiRoutes := api.Group("/ai")
		{
			aiRoutes.POST("/scan", ocrH.ScanReceipt)
			aiRoutes.GET("/advice", func(c *gin.Context) {
				advice, err := aiSvc.GetFinancialAdvice(c.Request.Context())
				if err != nil {
					c.JSON(500, gin.H{"error": "AI failure: " + err.Error()})
					return
				}
				c.JSON(200, gin.H{"advice": advice})
			})
		}

		// Goals
		goalRoutes := api.Group("/goals")
		{
			goalRoutes.POST("/", goalH.Create)
			goalRoutes.GET("/", goalH.List)
			goalRoutes.POST("/:id/contribute", goalH.AddContribution)
		}

		// Budgets
		budgetRoutes := api.Group("/budgets")
		{
			budgetRoutes.POST("/", budgetH.Create)
			budgetRoutes.GET("/status", budgetH.Status)
		}

		// Invoices (Freelance Module)
		invoiceRoutes := api.Group("/invoices")
		{
			invoiceRoutes.POST("/", invoiceH.Create)
			invoiceRoutes.GET("/:id/pdf", invoiceH.GeneratePDF)
		}

		// Analytics & Sync
		analyticsRoutes := api.Group("/data")
		{
			analyticsRoutes.GET("/sync", analyticsH.SynchronizeChanges)
			analyticsRoutes.GET("/reports/monthly", analyticsH.GetMonthlyReport)
		}
	}

	// Start App
	log.Printf("Starting Server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}
