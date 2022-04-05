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

type Order struct {
	Id             int         `json:"id" db:"id"`
	Fullname       string      `json:"fullname" db:"fullname"`
	Phone          string      `json:"phone" db:"phone"`
	Address        string      `json:"address" db:"address"`
	ProofOfPayment string      `json:"proof_of_payment" db:"proof_of_payment"`
	Status         string      `json:"status" db:"status"`
	NoReceipt      null.String `json:"no_receipt" db:"no_receipt"`
	TotalAmount    int         `json:"total_amount" db:"total_amount"`
	UserId         int         `json:"user_id" db:"user_id"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
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
