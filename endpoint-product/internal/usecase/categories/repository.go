package categories

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/auth"
	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/categories"
)

type categoriesRepo interface {
	GetCategoryByName(ctx context.Context, name string) (*categoriesentity.Category, error)
	GetCategoryById(ctx context.Context, id int) (*categoriesentity.Category, error)
	Insert(ctx context.Context, payload *categoriesentity.JsonCreateUpdateSchema) int
	GetAllCategoryPaginate(ctx context.Context,
		payload *categoriesentity.QueryParamAllCategorySchema) (*categoriesentity.CategoryPaginate, error)
	Update(ctx context.Context, payload *categoriesentity.JsonCreateUpdateSchema) error
	Delete(ctx context.Context, categoryId int) error
}

type authRepo interface {
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
}
