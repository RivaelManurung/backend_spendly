package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base holds common fields for all GORM models
type Base struct {
	ID        string         `gorm:"type:text;primary_key;" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hooks to auto-generate UUIDs before saving to DB
func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}

type Category struct {
	Base
	Label string `gorm:"type:text;not null" json:"label"`
	Icon  string `gorm:"type:text" json:"icon"`
	Color string `gorm:"type:text" json:"color"`         // HEX
	Type  string `gorm:"type:text;not null" json:"type"` // 'income', 'expense'
}

type Transaction struct {
	Base
	Title       string    `gorm:"type:text;not null" json:"title"`
	Amount      float64   `gorm:"type:real;not null" json:"amount"`
	Date        time.Time `gorm:"index;not null" json:"date"`
	CategoryID  string    `gorm:"type:text;index;not null" json:"category_id"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Type        string    `gorm:"type:text;not null" json:"type"` // 'income', 'expense', 'goal'
	Note        string    `gorm:"type:text" json:"note"`
	IsRecurring bool      `gorm:"type:boolean" json:"is_recurring"` 
	DeviceID    string    `gorm:"type:text;index" json:"device_id"`
}

type Goal struct {
	Base
	Title         string    `gorm:"type:text;not null" json:"title"`
	TargetAmount  float64   `gorm:"type:real;not null" json:"target_amount"`
	CurrentAmount float64   `gorm:"type:real;default:0" json:"current_amount"`
	Icon          string    `gorm:"type:text" json:"icon"`
	Color         string    `gorm:"type:text" json:"color"`
	TargetDate    time.Time `json:"target_date"`
}

type GoalContribution struct {
	Base
	GoalID        string      `gorm:"type:text;index;not null" json:"goal_id"`
	Goal          Goal        `gorm:"foreignKey:GoalID" json:"-"`
	TransactionID string      `gorm:"type:text;uniqueIndex;not null" json:"transaction_id"`
	Transaction   Transaction `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
	Amount        float64     `gorm:"type:real;not null" json:"amount"`
	Date          time.Time   `json:"date"`
	Note          string      `gorm:"type:text" json:"note"`
}

type Budget struct {
	Base
	CategoryID string   `gorm:"type:text;index;not null" json:"category_id"`
	Category   Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Amount     float64  `gorm:"type:real;not null" json:"amount"`
	Period     string   `gorm:"type:text;not null" json:"period"` // 'monthly', 'weekly'
	StartDate  time.Time `json:"start_date"`
}

type Invoice struct {
	Base
	ClientName  string    `gorm:"type:text;not null" json:"client_name"`
	ClientEmail string    `gorm:"type:text" json:"client_email"`
	Amount      float64   `gorm:"type:real;not null" json:"amount"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `gorm:"type:text;not null" json:"status"` // 'draft', 'sent', 'paid', 'overdue'
	Items       string    `gorm:"type:json" json:"items"`           // JSON array of items
}
