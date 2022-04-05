package orders

import (
	"context"
	"fmt"

	ordersentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/orders"
	"github.com/creent-production/cdk-go/pagination"

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

func (r *RepoOrders) GetAllOrderPaginate(ctx context.Context,
	payload *ordersentity.QueryParamAllOrderSchema, isAdmin bool) (*ordersentity.OrderPaginate, error) {

	var results ordersentity.OrderPaginate

	query := r.queries["getOrderByDynamic"] + " WHERE 1=1"
	if !isAdmin {
		query += ` AND user_id = :user_id`
	}
	if len(payload.Status) > 0 {
		query += ` AND status = :status`
	}
	query += ` ORDER BY id DESC`

	// pagination
	var count struct{ Total int }
	stmt_count, _ := r.db.PrepareNamedContext(ctx, fmt.Sprintf("SELECT count(*) AS total FROM (%s) AS anon_1", query))
	err := stmt_count.GetContext(ctx, &count, payload)
	if err != nil {
		return &results, err
	}
	payload.Offset = (payload.Page - 1) * payload.PerPage

	// results
	query += ` LIMIT :per_page OFFSET :offset`
	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	err = stmt.SelectContext(ctx, &results.Data, payload)
	if err != nil {
		return &results, err
	}

	paginate := pagination.Paginate{Page: payload.Page, PerPage: payload.PerPage, Total: count.Total}
	results.Total = paginate.Total
	results.NextNum = paginate.NextNum()
	results.PrevNum = paginate.PrevNum()
	results.Page = paginate.Page
	results.IterPages = paginate.IterPages()

	return &results, nil
}

func (r *RepoOrders) GetAllOrderItems(ctx context.Context, orderId int) ([]ordersentity.OrderItemProduct, error) {
	var results []ordersentity.OrderItemProduct

	query := `
	SELECT
	transaction.order_items.id as order_items_id,
	transaction.order_items.notes as order_items_notes,
	transaction.order_items.qty as order_items_qty,
	transaction.order_items.price as order_items_price,
	transaction.order_items.product_id as order_items_product_id,
	transaction.order_items.created_at as order_items_created_at,
	transaction.order_items.updated_at as order_items_updated_at,
	product.products.name as product_name,
	product.products.slug as product_slug,
	product.products.image as product_image
FROM transaction.order_items
INNER JOIN product.products ON product.products.id = transaction.order_items.product_id
WHERE transaction.order_items.order_id = :order_id ORDER BY transaction.order_items.id DESC
	`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	err := stmt.SelectContext(ctx, &results, ordersentity.OrderItem{OrderId: orderId})
	if err != nil {
		return results, err
	}

	return results, nil
}
