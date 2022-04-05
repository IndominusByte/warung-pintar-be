package products

import "time"

type Product struct {
	Id          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	Image       string    `json:"image" db:"image"`
	Price       int       `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	CategoryId  int       `json:"category_id" db:"category_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
