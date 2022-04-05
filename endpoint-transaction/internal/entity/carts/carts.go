package carts

import "gopkg.in/guregu/null.v4"

type JsonCreateUpdateSchema struct {
	Operation string `json:"operation" validate:"required,oneof=create update"`
	ProductId int    `json:"product_id" validate:"required,gte=1"`
	Notes     string `json:"notes" validate:"omitempty,min=3,max=100"`
	Qty       int    `json:"qty" validate:"required,gte=1"`
}

type Cart struct {
	Id        int         `json:"id" db:"id"`
	Notes     null.String `json:"notes" db:"notes"`
	Qty       int         `json:"qty" db:"qty"`
	UserId    int         `json:"user_id" db:"user_id"`
	ProductId int         `json:"product_id" db:"product_id"`
}
