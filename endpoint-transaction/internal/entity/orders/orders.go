package orders

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type FormCreateSchema struct {
	Fullname       string `schema:"fullname" validate:"required,min=3,max=100" db:"fullname"`
	Phone          string `schema:"phone" validate:"required,phone=id" db:"phone"`
	Address        string `schema:"address" validate:"required,min=5" db:"address"`
	ProofOfPayment string `schema:"-" db:"proof_of_payment"`
	TotalAmount    int    `schema:"-" db:"total_amount"`
	UserId         int    `schema:"-" db:"user_id"`
}

type QueryParamAllOrderSchema struct {
	UserId  int    `schema:"-" db:"user_id"`
	Page    int    `schema:"page" validate:"required,gte=1"`
	PerPage int    `schema:"per_page" validate:"required,gte=1" db:"per_page"`
	Status  string `schema:"status" validate:"omitempty,oneof='ongoing' 'reject' 'on the way' 'success'" db:"status"`
	Offset  int    `schema:"-" db:"offset"`
}

type Order struct {
	Id             int                `json:"id" db:"id"`
	Fullname       string             `json:"fullname" db:"fullname"`
	Phone          string             `json:"phone" db:"phone"`
	Address        string             `json:"address" db:"address"`
	ProofOfPayment string             `json:"proof_of_payment" db:"proof_of_payment"`
	Status         string             `json:"status" db:"status"`
	NoReceipt      null.String        `json:"no_receipt" db:"no_receipt"`
	TotalAmount    int                `json:"total_amount" db:"total_amount"`
	UserId         int                `json:"user_id" db:"user_id"`
	OrderItems     []OrderItemProduct `json:"order_items"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	Id        int         `json:"id" db:"id"`
	Notes     null.String `json:"notes" db:"notes"`
	Qty       int         `json:"qty" db:"qty"`
	Price     int         `json:"price" db:"price"`
	ProductId int         `json:"product_id" db:"product_id"`
	OrderId   int         `json:"order_id" db:"order_id"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItemProduct struct {
	OrderItemsId        int         `json:"order_items_id" db:"order_items_id"`
	OrderItemsNotes     null.String `json:"order_items_notes" db:"order_items_notes"`
	OrderItemsQty       int         `json:"order_items_qty" db:"order_items_qty"`
	OrderItemsPrice     int         `json:"order_items_price" db:"order_items_price"`
	OrderItemsProductId int         `json:"order_items_product_id" db:"order_items_product_id"`
	OrderItemsCreatedAt time.Time   `json:"order_items_created_at" db:"order_items_created_at"`
	OrderItemsUpdatedAt time.Time   `json:"order_items_updated_at" db:"order_items_updated_at"`
	ProductName         string      `json:"product_name" db:"product_name"`
	ProductSlug         string      `json:"product_slug" db:"product_slug"`
	ProductImage        string      `json:"product_image" db:"product_image"`
}

type OrderPaginate struct {
	Data      []Order    `json:"data"`
	Total     int        `json:"total"`
	NextNum   null.Int   `json:"next_num"`
	PrevNum   null.Int   `json:"prev_num"`
	Page      int        `json:"page"`
	IterPages []null.Int `json:"iter_pages"`
}
