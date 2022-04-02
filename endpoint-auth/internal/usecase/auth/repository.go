package auth

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/entity/auth"
)

type authRepo interface {
	IsPasswordSameAsHash(ctx context.Context, hash, password []byte) bool
	GetUserByEmail(ctx context.Context, email string) (*authentity.User, error)
	GetUserById(ctx context.Context, userId int) (*authentity.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*authentity.User, error)
	GetUserConfirmById(ctx context.Context, id string) (*authentity.UserConfirm, error)
	GetUserConfirmByUserId(ctx context.Context, id int) (*authentity.UserConfirm, error)
	GetPasswordResetById(ctx context.Context, id string) (*authentity.PasswordReset, error)
	GetPasswordResetByEmail(ctx context.Context, email string) (*authentity.PasswordReset, error)
	InsertUser(ctx context.Context, payload *authentity.JsonRegisterSchema) int
	UpdateUser(ctx context.Context, payload *authentity.User) error
	InsertUserConfirm(ctx context.Context, user_id int) string
	InsertPasswordReset(ctx context.Context, payload *authentity.JsonEmailSchema) string
	DeletePasswordReset(ctx context.Context, id string) error
	SetUserConfirmActivatedTrue(ctx context.Context, id string) error
	GenerateUserConfirmResendExpired(ctx context.Context, id string) error
	GeneratePasswordResetResendExpired(ctx context.Context, id string) error
}
