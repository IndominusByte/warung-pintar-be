package orders

import (
	"context"

	ordersentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/orders"

	"github.com/jmoiron/sqlx"
)

type RepoOrders struct {
	db      *sqlx.DB
	queries map[string]string
	execs   map[string]string
}

var queries = map[string]string{
	"getOrderByDynamic": `SELECT id, fullname, phone, address, proof_of_payment, status, no_receipt, total_amount, user_id, created_at, updated_at FROM transaction.orders`,
}
var execs = map[string]string{
	"insertOrder":     `INSERT INTO transaction.orders (fullname, phone, address, proof_of_payment, total_amount, user_id) VALUES (:fullname, :phone, :address, :proof_of_payment, :total_amount, :user_id) RETURNING id`,
	"insertOrderItem": `INSERT INTO transaction.order_items (notes, qty, price, product_id, order_id) VALUES (:notes, :qty, :price, :product_id, :order_id) RETURNING id`,
}

func New(db *sqlx.DB) (*RepoOrders, error) {
	rp := &RepoOrders{
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
func (r *RepoOrders) Validate() error {
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

func (r *RepoOrders) GetOrderById(ctx context.Context, orderId int) (*ordersentity.Order, error) {
	var t ordersentity.Order
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getOrderByDynamic"]+" WHERE id = :id")

	return &t, stmt.GetContext(ctx, &t, ordersentity.Order{Id: orderId})
}

func (r *RepoOrders) Insert(ctx context.Context, payload *ordersentity.FormCreateSchema) int {
	var id int
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertOrder"])
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoOrders) InsertItem(ctx context.Context, payload *ordersentity.OrderItem) int {
	var id int
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertOrderItem"])
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoOrders) UpdateOrder(ctx context.Context, payload *ordersentity.Order) error {
	query := `UPDATE transaction.orders SET updated_at=CURRENT_TIMESTAMP`
	if len(payload.Status) > 0 {
		query += `, status=:status`
	}
	if len(payload.NoReceipt.String) > 0 {
		query += `, no_receipt=:no_receipt`
	}
	query += ` WHERE id = :id`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	_, err := stmt.ExecContext(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}
