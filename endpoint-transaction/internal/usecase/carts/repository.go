package carts

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/auth"
)

type cartsRepo interface {
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}
