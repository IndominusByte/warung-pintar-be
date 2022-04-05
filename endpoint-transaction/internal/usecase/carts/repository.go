package carts

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/auth"
	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/products"
	"github.com/gomodule/redigo/redis"
)

type cartsRepo interface {
	GetCartByUserIdAndProductId(ctx context.Context, userId, productId int) (*cartsentity.Cart, error)
	GetAllCarts(ctx context.Context,
		payload *cartsentity.QueryParamAllCartSchema) ([]cartsentity.CartProduct, error)
	Insert(ctx context.Context, payload *cartsentity.Cart) int
	Update(ctx context.Context, payload *cartsentity.Cart) error
	Delete(ctx context.Context, payload *cartsentity.JsonMultipleSchema) error
	MoveItemToPayment(ctx context.Context, redisCli *redis.Pool, payload *cartsentity.JsonMultipleSchema)
	ItemInPayment(ctx context.Context, redisCli *redis.Pool, userId int) ([]cartsentity.CartProduct, error)
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}

type productsRepo interface {
	GetProductById(ctx context.Context, id int) (*productsentity.Product, error)
}
