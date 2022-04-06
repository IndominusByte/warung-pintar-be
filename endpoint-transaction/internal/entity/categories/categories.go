package categories

import (
	"time"
)

type JsonCreateUpdateSchema struct {
	Id   int    `db:"id"`
	Name string `validate:"required,min=3,max=100" db:"name"`
}

type Category struct {
	Id        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
