package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	// Import driver pgx agar kompatibel dengan stdlib "database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Database membungkus sqlx.DB agar nanti mudah disuntikkan methode kustom atau metrics
type Database struct {
	*sqlx.DB
}

// ConfigDB menampung konfigurasi opsional untuk koneksi ke postgres
type ConfigDB struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewPostgresDB menginisialisasi dan mengetes koneksi ke Postgres (level production).
func NewPostgresDB(ctx context.Context, cfg ConfigDB) (*Database, error) {
	// Menghubungkan via driver "pgx" melalui interface database/sql standar (dibungkus sqlx)
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to postgres: %w", err)
	}

	// ---------------------------------------------------------
	// Production Connection Pool Tunings (bisa disesuaikan config)
	// ---------------------------------------------------------

	// SetMaxOpenConns menetapkan jumlah maksimal open connections.
	// Jangan terlalu besar, disarankan ≤ jumlah core DB. (default kita pakai 25 jika 0)
	maxOpen := cfg.MaxOpenConns
	if maxOpen == 0 {
		maxOpen = 25
	}
	db.SetMaxOpenConns(maxOpen)

	// SetMaxIdleConns sebaiknya sama dengan SetMaxOpenConns pada mayoritas API
	// agar koneksi tidak terus ditutup-buka yang menyebabkan latensi membesar.
	maxIdle := cfg.MaxIdleConns
	if maxIdle == 0 {
		maxIdle = 25
	}
	db.SetMaxIdleConns(maxIdle)

	// SetConnMaxLifetime mencegah koneksi bertahan selamanya untuk menghindari
	// stale connections (Firewall drop/DB proxy tcp drop).
	lifeTime := cfg.ConnMaxLifetime
	if lifeTime == 0 {
		lifeTime = 15 * time.Minute
	}
	db.SetConnMaxLifetime(lifeTime)

	// SetConnMaxIdleTime mencegah membebani server saat traffic sepi
	idleTime := cfg.ConnMaxIdleTime
	if idleTime == 0 {
		idleTime = 5 * time.Minute
	}
	db.SetConnMaxIdleTime(idleTime)

	// Verifikasi akhir dengan ping (memastikan DB benar-benar alive)
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed after connection: %w", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL database")

	return &Database{db}, nil
}
