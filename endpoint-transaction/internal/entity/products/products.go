package products

import "time"

type FormCreateUpdateSchema struct {
	Id          int    `schema:"-" db:"id"`
	Name        string `schema:"name" validate:"required,min=3,max=100" db:"name"`
	Slug        string `schema:"-" db:"slug"`
	Description string `schema:"description" validate:"required,min=5" db:"description"`
	Image       string `schema:"-" db:"image"`
	Price       int    `schema:"price" validate:"required,min=1" db:"price"`
	Stock       int    `schema:"stock" validate:"required,min=1" db:"stock"`
	CategoryId  int    `schema:"category_id" validate:"required,min=1" db:"category_id"`
}

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
