package carts

import "gopkg.in/guregu/null.v4"

type JsonCreateUpdateSchema struct {
	Operation string `json:"operation" validate:"required,oneof=create update"`
	ProductId int    `json:"product_id" validate:"required,gte=1"`
	Notes     string `json:"notes" validate:"omitempty,min=3,max=100"`
	Qty       int    `json:"qty" validate:"required,gte=1"`
}

type JsonMultipleSchema struct {
	UserId int   `db:"user_id"`
	ListId []int `json:"list_id" validate:"required,min=1,unique,dive,required,min=1"`
}

type QueryParamAllCartSchema struct {
	UserId int    `schema:"-" db:"user_id"`
	Stock  string `schema:"stock" validate:"omitempty,oneof=empty ready"`
}

type Cart struct {
	Id        int         `json:"id" db:"id"`
	Notes     null.String `json:"notes" db:"notes"`
	Qty       int         `json:"qty" db:"qty"`
	UserId    int         `json:"user_id" db:"user_id"`
	ProductId int         `json:"product_id" db:"product_id"`
}

type CartProduct struct {
	CartId        int         `json:"cart_id" db:"cart_id"`
	CartNotes     null.String `json:"cart_notes" db:"cart_notes"`
	CartQty       int         `json:"cart_qty" db:"cart_qty"`
	CartUserId    int         `json:"cart_user_id" db:"cart_user_id"`
	CartProductId int         `json:"cart_product_id" db:"cart_product_id"`
	ProductName   string      `json:"product_name" db:"product_name"`
	ProductSlug   string      `json:"product_slug" db:"product_slug"`
	ProductImage  string      `json:"product_image" db:"product_image"`
	ProductPrice  int         `json:"product_price" db:"product_price"`
	ProductStock  int         `json:"product_stock" db:"product_stock"`
}
