package products

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/auth"
)

type productsRepo interface {
}

type categoriesRepo interface {
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}
