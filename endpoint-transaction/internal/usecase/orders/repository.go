package orders

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/auth"
	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	ordersentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/orders"
	"github.com/gomodule/redigo/redis"
)

type ordersRepo interface {
	GetOrderById(ctx context.Context, orderId int) (*ordersentity.Order, error)
	UpdateOrder(ctx context.Context, payload *ordersentity.Order) error
	Insert(ctx context.Context, payload *ordersentity.FormCreateSchema) int
	InsertItem(ctx context.Context, payload *ordersentity.OrderItem) int
	GetAllOrderPaginate(ctx context.Context,
		payload *ordersentity.QueryParamAllOrderSchema, isAdmin bool) (*ordersentity.OrderPaginate, error)
	GetAllOrderItems(ctx context.Context, orderId int) ([]ordersentity.OrderItemProduct, error)
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}

type cartsRepo interface {
	Delete(ctx context.Context, payload *cartsentity.JsonMultipleSchema) error
	ItemInPayment(ctx context.Context, redisCli *redis.Pool, userId int) ([]cartsentity.CartProduct, error)
}
