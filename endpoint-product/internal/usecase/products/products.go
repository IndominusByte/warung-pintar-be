package products

type ProductsUsecase struct {
	productsRepo   productsRepo
	categoriesRepo categoriesRepo
	authRepo       authRepo
}

func NewProductsUsecase(productRepo productsRepo, categoryRepo categoriesRepo, authRepo authRepo) *ProductsUsecase {
	return &ProductsUsecase{
		productsRepo:   productRepo,
		categoriesRepo: categoryRepo,
		authRepo:       authRepo,
	}
}
