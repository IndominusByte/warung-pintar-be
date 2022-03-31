package auth

type AuthUsecase struct {
	authRepo authRepo
}

func NewAuthUsecase(authRepo authRepo) *AuthUsecase {
	return &AuthUsecase{
		authRepo: authRepo,
	}
}
