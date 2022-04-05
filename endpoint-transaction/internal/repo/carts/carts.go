package carts

import (
	"context"

	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	"github.com/jmoiron/sqlx"
)

type RepoCarts struct {
	db      *sqlx.DB
	queries map[string]string
	execs   map[string]string
}

var queries = map[string]string{
	"getCartByDynamic": `SELECT id, notes, qty, user_id, product_id FROM transaction.carts`,
}
var execs = map[string]string{
	"updateCart": `UPDATE transaction.carts SET qty=:qty, user_id=:user_id, product_id=:product_id, notes=:notes WHERE id = :id`,
}

func New(db *sqlx.DB) (*RepoCarts, error) {
	rp := &RepoCarts{
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
func (r *RepoCarts) Validate() error {
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

func (r *RepoCarts) GetCartByUserIdAndProductId(ctx context.Context, userId, productId int) (*cartsentity.Cart, error) {
	var t cartsentity.Cart
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getCartByDynamic"]+" WHERE user_id = :user_id AND product_id = :product_id")

	return &t, stmt.GetContext(ctx, &t, cartsentity.Cart{UserId: userId, ProductId: productId})
}

func (r *RepoCarts) Insert(ctx context.Context, payload *cartsentity.Cart) int {
	var id int

	query := `INSERT INTO transaction.carts (qty, user_id, product_id`

	if len(payload.Notes.String) > 0 {
		query += `, notes`
	}

	query += `) VALUES (:qty, :user_id, :product_id`

	if len(payload.Notes.String) > 0 {
		query += `, :notes`
	}

	query += `) RETURNING id`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoCarts) Update(ctx context.Context, payload *cartsentity.Cart) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["updateCart"])
	_, err := stmt.ExecContext(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}
