package entities

import (
	"time"
)

type StockMovement struct {
	ID            string     `json:"id,omitempty"`
	ProductID     string     `json:"product_id"`
	Type          string     `json:"type"`
	Quantity      int        `json:"quantity"`
	PreviousStock int        `json:"previous_stock"`
	CurrentStock  int        `json:"current_stock"`
	Reason        *string    `json:"reason,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}
