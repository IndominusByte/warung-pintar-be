package auth

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/entity/auth"
)

type authRepo interface {
	IsPasswordSameAsHash(ctx context.Context, hash, password []byte) bool
	GetUserByEmail(ctx context.Context, email string) (*authentity.User, error)
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
	GetUserConfirmById(ctx context.Context, id string) (*authentity.UserConfirm, error)
	GetUserConfirmByUserId(ctx context.Context, id int) (*authentity.UserConfirm, error)
	InsertUser(ctx context.Context, payload *authentity.JsonRegisterSchema) int
	InsertUserConfirm(ctx context.Context, user_id int) string
	SetUserConfirmActivatedTrue(ctx context.Context, id string) error
	GenerateUserConfirmResendExpired(ctx context.Context, id string) error
}
