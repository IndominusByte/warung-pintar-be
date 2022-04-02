package auth

import (
	"context"

	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/entity/auth"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"
)

type RepoAuth struct {
	db      *sqlx.DB
	queries map[string]string
	execs   map[string]string
}

var queries = map[string]string{
	"getUserByDynamic":          `SELECT id, fullname, email, password, phone, address, role, avatar, created_at, updated_at FROM account.users`,
	"getUserConfirmByDynamic":   `SELECT id, activated, resend_expired, user_id FROM account.confirmation_users`,
	"getPasswordResetByDynamic": `SELECT id, email, resend_expired, created_at FROM account.password_resets`,
}
var execs = map[string]string{
	"insertUser":                         `INSERT INTO account.users (email, password) VALUES (:email, :password) RETURNING id`,
	"insertUserConfirm":                  `INSERT INTO account.confirmation_users (user_id, resend_expired) VALUES (:id, (CURRENT_TIMESTAMP + INTERVAL '5 minute')) RETURNING id`,
	"insertPasswordReset":                `INSERT INTO account.password_resets (email, resend_expired) VALUES (:email, (CURRENT_TIMESTAMP + INTERVAL '5 minute')) RETURNING id`,
	"setUserConfirmActivatedTrue":        `UPDATE account.confirmation_users SET activated=true WHERE id = :id`,
	"generateUserConfirmResendExpired":   `UPDATE account.confirmation_users SET resend_expired=(CURRENT_TIMESTAMP + INTERVAL '5 minute') WHERE id = :id`,
	"generatePasswordResetResendExpired": `UPDATE account.password_resets SET resend_expired=(CURRENT_TIMESTAMP + INTERVAL '5 minute') WHERE id = :id`,
	"deletePasswordReset":                `DELETE FROM account.password_resets WHERE id = :id`,
}

func New(db *sqlx.DB) (*RepoAuth, error) {
	rp := &RepoAuth{
		db:      db,
		queries: queries,
		execs:   execs,
	}

	err := rp.Validate()
	if err != nil {
		return nil, err
	}

	return rp, nil
}

// Validate will validate sql query to db
func (r *RepoAuth) Validate() error {
	for _, q := range r.queries {
		_, err := r.db.PrepareNamedContext(context.Background(), q)
		if err != nil {
			return err
		}
	}

	for _, e := range r.execs {
		_, err := r.db.PrepareNamedContext(context.Background(), e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RepoAuth) IsPasswordSameAsHash(ctx context.Context, hash, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		return false
	}
	return true
}

func (r *RepoAuth) GetUserByEmail(ctx context.Context, email string) (*authentity.User, error) {
	var t authentity.User
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getUserByDynamic"]+" WHERE email = :email")

	return &t, stmt.GetContext(ctx, &t, authentity.User{Email: email})
}

func (r *RepoAuth) GetUserById(ctx context.Context, userId int) (*authentity.User, error) {
	var t authentity.User
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getUserByDynamic"]+" WHERE id = :id")

	return &t, stmt.GetContext(ctx, &t, authentity.User{Id: userId})
}

func (r *RepoAuth) GetUserByPhone(ctx context.Context, phone string) (*authentity.User, error) {
	var t authentity.User
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getUserByDynamic"]+" WHERE phone = :phone")

	return &t, stmt.GetContext(ctx, &t, authentity.User{Phone: null.StringFrom(phone)})
}

func (r *RepoAuth) GetUserConfirmById(ctx context.Context, id string) (*authentity.UserConfirm, error) {
	var t authentity.UserConfirm
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getUserConfirmByDynamic"]+" WHERE id = :id")

	return &t, stmt.GetContext(ctx, &t, authentity.UserConfirm{Id: id})
}

func (r *RepoAuth) GetUserConfirmByUserId(ctx context.Context, id int) (*authentity.UserConfirm, error) {
	var t authentity.UserConfirm
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getUserConfirmByDynamic"]+" WHERE user_id = :id")

	return &t, stmt.GetContext(ctx, &t, authentity.User{Id: id})
}

func (r *RepoAuth) GetPasswordResetById(ctx context.Context, id string) (*authentity.PasswordReset, error) {
	var t authentity.PasswordReset

	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getPasswordResetByDynamic"]+" WHERE id = :id")

	return &t, stmt.GetContext(ctx, &t, authentity.PasswordReset{Id: id})
}

func (r *RepoAuth) GetPasswordResetByEmail(ctx context.Context, email string) (*authentity.PasswordReset, error) {
	var t authentity.PasswordReset

	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getPasswordResetByDynamic"]+" WHERE email = :email")

	return &t, stmt.GetContext(ctx, &t, authentity.PasswordReset{Email: email})
}

func (r *RepoAuth) InsertUser(ctx context.Context, payload *authentity.JsonRegisterSchema) int {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	payload.Password = string(hashedPassword)

	var id int
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertUser"])
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoAuth) UpdateUser(ctx context.Context, payload *authentity.User) error {
	query := `UPDATE account.users SET updated_at=CURRENT_TIMESTAMP`
	if len(payload.Fullname.String) > 0 {
		query += `, fullname=:fullname`
	}
	if len(payload.Email) > 0 {
		query += `, email=:email`
	}
	if len(payload.Password) > 0 {
		// hash password
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		payload.Password = string(hashedPassword)

		query += `, password=:password`
	}
	if len(payload.Phone.String) > 0 {
		query += `, phone=:phone`
	}
	if len(payload.Address.String) > 0 {
		query += `, address=:address`
	}
	if len(payload.Role) > 0 {
		query += `, role=:role`
	}
	if len(payload.Avatar) > 0 {
		query += `, avatar=:avatar`
	}
	query += ` WHERE id = :id`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	_, err := stmt.ExecContext(ctx, payload)
	if err != nil {
		return err
	}
	return nil

}

func (r *RepoAuth) InsertUserConfirm(ctx context.Context, user_id int) string {
	var id string
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertUserConfirm"])
	stmt.QueryRowxContext(ctx, authentity.User{Id: user_id}).Scan(&id)

	return id
}

func (r *RepoAuth) InsertPasswordReset(ctx context.Context, payload *authentity.JsonEmailSchema) string {
	var id string
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertPasswordReset"])
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoAuth) DeletePasswordReset(ctx context.Context, id string) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["deletePasswordReset"])
	_, err := stmt.ExecContext(ctx, authentity.PasswordReset{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoAuth) SetUserConfirmActivatedTrue(ctx context.Context, id string) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["setUserConfirmActivatedTrue"])
	_, err := stmt.ExecContext(ctx, authentity.UserConfirm{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoAuth) GenerateUserConfirmResendExpired(ctx context.Context, id string) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["generateUserConfirmResendExpired"])
	_, err := stmt.ExecContext(ctx, authentity.UserConfirm{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoAuth) GeneratePasswordResetResendExpired(ctx context.Context, id string) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["generatePasswordResetResendExpired"])
	_, err := stmt.ExecContext(ctx, authentity.PasswordReset{Id: id})
	if err != nil {
		return err
	}
	return nil
}
