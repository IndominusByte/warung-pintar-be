package carts

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/auth"
	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/products"
)

type cartsRepo interface {
	GetCartByUserIdAndProductId(ctx context.Context, userId, productId int) (*cartsentity.Cart, error)
	Insert(ctx context.Context, payload *cartsentity.Cart) int
	Update(ctx context.Context, payload *cartsentity.Cart) error
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}

type productsRepo interface {
	GetProductById(ctx context.Context, id int) (*productsentity.Product, error)
}
