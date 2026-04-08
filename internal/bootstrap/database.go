package bootstrap

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/spendly/backend/internal/config"
	"github.com/spendly/backend/internal/domain"
)

func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if cfg.DBType == "postgres" {
		dialector = postgres.Open(cfg.DBDSN)
		log.Printf("Connecting to PostgreSQL: %s", cfg.DBDSN)
	} else {
		dialector = sqlite.Open(cfg.DBDSN)
		log.Printf("Connecting to SQLite: %s", cfg.DBDSN)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Show SQL in terminal
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	log.Printf("%s Database connected successfully.", cfg.DBType)

	// Auto Migrate schemas (Single-user mode)
	err = db.AutoMigrate(
		&domain.Category{},
		&domain.Transaction{},
		&domain.Goal{},
		&domain.GoalContribution{},
		&domain.Budget{},
		&domain.Invoice{},
	)

	if err != nil {
		log.Fatalf("Failed to execute GORM migrations: %v", err)
		return nil, err
	}

	log.Println("Database schemas auto-migrated via GORM.")
	return db, nil
}
