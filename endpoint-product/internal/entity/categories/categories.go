package categories

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type JsonCreateUpdateSchema struct {
	Id   int    `db:"id"`
	Name string `validate:"required,min=3,max=100" db:"name"`
}

type QueryParamAllCategorySchema struct {
	Page    int    `schema:"page" validate:"required,gte=1"`
	PerPage int    `schema:"per_page" validate:"required,gte=1" db:"per_page"`
	Q       string `schema:"q" db:"q"`
	Offset  int    `schema:"-" db:"offset"`
}

type Category struct {
	Id        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CategoryPaginate struct {
	Data      []Category `json:"data"`
	Total     int        `json:"total"`
	NextNum   null.Int   `json:"next_num"`
	PrevNum   null.Int   `json:"prev_num"`
	Page      int        `json:"page"`
	IterPages []null.Int `json:"iter_pages"`
}
