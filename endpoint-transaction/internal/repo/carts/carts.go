package carts

import (
	"context"
	"fmt"
	"strings"

	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	"github.com/creent-production/cdk-go/parser"
	"github.com/gomodule/redigo/redis"
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
	"deleteCart": `DELETE FROM transaction.carts WHERE user_id = :user_id`,
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

func (r *RepoCarts) GetAllCarts(ctx context.Context,
	payload *cartsentity.QueryParamAllCartSchema) ([]cartsentity.CartProduct, error) {
	var results []cartsentity.CartProduct

	query := `
SELECT
    transaction.carts.id as cart_id,
    transaction.carts.notes as cart_notes,
    transaction.carts.qty as cart_qty,
	transaction.carts.user_id as cart_user_id,
	transaction.carts.product_id as cart_product_id,
	product.products.name as product_name,
	product.products.slug as product_slug,
	product.products.image as product_image,
	product.products.price as product_price,
	product.products.stock as product_stock
FROM
    transaction.carts
INNER JOIN product.products ON product.products.id = transaction.carts.product_id
WHERE transaction.carts.user_id = :user_id
	`

	if payload.Stock == "ready" {
		query += ` AND product.products.stock > 0`
	}
	if payload.Stock == "empty" {
		query += ` AND product.products.stock < 1`
	}
	query += ` ORDER BY transaction.carts.id DESC`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	err := stmt.SelectContext(ctx, &results, payload)
	if err != nil {
		return results, err
	}

	return results, nil
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

func (r *RepoCarts) Delete(ctx context.Context, payload *cartsentity.JsonMultipleSchema) error {
	query := r.execs["deleteCart"]
	query += fmt.Sprintf(` AND id IN (%s)`, strings.Join(parser.ParseSliceIntToSliceStr(payload.ListId), ","))

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	_, err := stmt.ExecContext(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoCarts) DeleteByUserId(ctx context.Context, userId int) error {
	query := r.execs["deleteCart"]
	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	_, err := stmt.ExecContext(ctx, cartsentity.JsonMultipleSchema{UserId: userId})
	if err != nil {
		return err
	}
	return nil

}

func (r *RepoCarts) MoveItemToPayment(ctx context.Context, redisCli *redis.Pool, payload *cartsentity.JsonMultipleSchema) {
	var (
		results []cartsentity.Cart
		data    []int
	)

	query := r.queries["getCartByDynamic"]
	query += ` WHERE user_id = :user_id`
	query += fmt.Sprintf(` AND id IN (%s)`, strings.Join(parser.ParseSliceIntToSliceStr(payload.ListId), ","))

	stmt, _ := r.db.PrepareNamedContext(ctx, query)

	stmt.SelectContext(ctx, &results, payload)
	for _, v := range results {
		data = append(data, v.Id)
	}

	conn := redisCli.Get()
	defer conn.Close()

	conn.Do("SETEX", fmt.Sprintf("checkout:%d", payload.UserId), 86400, strings.Join(parser.ParseSliceIntToSliceStr(data), ","))
}

func (r *RepoCarts) ItemInPayment(ctx context.Context, redisCli *redis.Pool, userId int) ([]cartsentity.CartProduct, error) {
	conn := redisCli.Get()
	defer conn.Close()

	item, _ := redis.String(conn.Do("GET", fmt.Sprintf("checkout:%d", userId)))
	if len(item) < 1 {
		item = "0"
	}

	var results []cartsentity.CartProduct

	query := `
SELECT
    transaction.carts.id as cart_id,
    transaction.carts.notes as cart_notes,
    transaction.carts.qty as cart_qty,
	transaction.carts.user_id as cart_user_id,
	transaction.carts.product_id as cart_product_id,
	product.products.name as product_name,
	product.products.slug as product_slug,
	product.products.image as product_image,
	product.products.price as product_price,
	product.products.stock as product_stock
FROM
    transaction.carts
INNER JOIN product.products ON product.products.id = transaction.carts.product_id
WHERE transaction.carts.user_id = :user_id AND product.products.stock > 0
	`

	query += fmt.Sprintf(` AND transaction.carts.id IN (%s)`, item)
	query += ` ORDER BY transaction.carts.id DESC`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	err := stmt.SelectContext(ctx, &results, cartsentity.Cart{UserId: userId})
	if err != nil {
		return results, err
	}

	return results, nil
}
