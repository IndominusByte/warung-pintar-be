package products

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/auth"
	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/categories"
	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/products"
)

type productsRepo interface {
	GetProductBySlug(ctx context.Context, slug string) (*productsentity.Product, error)
	GetProductById(ctx context.Context, id int) (*productsentity.Product, error)
	GetAllProductPaginate(ctx context.Context,
		payload *productsentity.QueryParamAllProductSchema) (*productsentity.ProductPaginate, error)
	Insert(ctx context.Context, payload *productsentity.FormCreateUpdateSchema) int
	Update(ctx context.Context, payload *productsentity.FormCreateUpdateSchema) error
	Delete(ctx context.Context, productId int) error
}

type categoriesRepo interface {
	GetCategoryById(ctx context.Context, id int) (*categoriesentity.Category, error)
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}
