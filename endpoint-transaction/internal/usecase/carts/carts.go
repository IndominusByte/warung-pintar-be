package carts

type CartsUsecase struct {
	cartsRepo cartsRepo
	authRepo  authRepo
}

func NewCartsUsecase(cartRepo cartsRepo, authRepo authRepo) *CartsUsecase {
	return &CartsUsecase{
		cartsRepo: cartRepo,
		authRepo:  authRepo,
	}
}
