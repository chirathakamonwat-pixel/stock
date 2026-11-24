package entities

import "time"

type Product struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	SKU        string     `json:"sku"`
	CategoryID *string    `json:"category_id"`
	Price      float64    `json:"price"`
	Stock      int        `json:"stock"`
	MinStock   int        `json:"min_stock"`
	Unit       string     `json:"unit"`
	CreatedAt  *time.Time `json:"created_at"`
}
