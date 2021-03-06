package auth

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type JsonRegisterSchema struct {
	Email           string `json:"email" validate:"required,email,min=3,max=100" db:"email"`
	Password        string `json:"password" validate:"required,min=6,max=100" db:"password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=6,max=100,eqfield=Password"`
}

type JsonEmailSchema struct {
	Email string `json:"email" validate:"required,email,min=3,max=100" db:"email"`
}

type JsonLoginSchema struct {
	Email    string `json:"email" validate:"required,email,min=3,max=100" db:"email"`
	Password string `json:"password" validate:"required,min=6,max=100" db:"password"`
}

type JsonPasswordOnlySchema struct {
	Password string `validate:"required,min=6,max=100"`
}

type JsonPasswordResetSchema struct {
	Email           string `json:"email" validate:"required,email,min=3,max=100" db:"email"`
	Password        string `json:"password" validate:"required,min=6,max=100" db:"password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=6,max=100,eqfield=Password"`
}

type JsonUpdatePasswordSchema struct {
	Password        string `json:"password" validate:"required,min=6,max=100" db:"password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=6,max=100,eqfield=Password"`
	OldPassword     string `json:"old_password" validate:"required,min=6,max=100"`
}

type JsonUpdateAccountSchema struct {
	Fullname string `validate:"required,min=3,max=100" db:"fullname"`
	Phone    string `json:"phone" validate:"required,phone=id" db:"phone"`
	Address  string `json:"address" validate:"required,min=5" db:"address"`
}

type User struct {
	Id        int         `json:"id" db:"id"`
	Fullname  null.String `json:"fullname" db:"fullname"`
	Email     string      `json:"email" db:"email"`
	Password  string      `json:"-" db:"password"`
	Phone     null.String `json:"phone" db:"phone"`
	Address   null.String `json:"address" db:"address"`
	Role      string      `json:"role" db:"role"`
	Avatar    string      `json:"avatar" db:"avatar"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

type UserConfirm struct {
	Id            string    `json:"id" db:"id"`
	Activated     bool      `json:"activated" db:"activated"`
	ResendExpired time.Time `json:"resend_expired" db:"resend_expired"`
	UserId        int       `json:"user_id" db:"user_id"`
}

type PasswordReset struct {
	Id            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	ResendExpired time.Time `json:"resend_expired" db:"resend_expired"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
