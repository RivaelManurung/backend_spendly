package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AIInsight struct {
	ID         int64           `db:"id" json:"id"`
	UserID     uuid.UUID       `db:"user_id" json:"user_id"`
	SnapshotID *int64          `db:"snapshot_id" json:"snapshot_id,omitempty"`
	Type       string          `db:"type" json:"type"` // e.g. "MONTHLY_SUMMARY", "ANOMALY", "SAVING_TIP"
	Title      string          `db:"title" json:"title"`
	Content    string          `db:"content" json:"content"`
	Priority   int             `db:"priority" json:"priority"`
	IsRead     bool            `db:"is_read" json:"is_read"`
	Metadata   json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at" json:"updated_at"`
}
