package products

import (
	"context"

	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/products"

	"github.com/jmoiron/sqlx"
)

type RepoProducts struct {
	db      *sqlx.DB
	queries map[string]string
	execs   map[string]string
}

var queries = map[string]string{
	"getProductByDynamic": `SELECT id, name, slug, description, image, price, stock, category_id, created_at, updated_at FROM product.products`,
}
var execs = map[string]string{
	"insertProduct": `INSERT INTO product.products (name, slug, description, image, price, stock, category_id) VALUES (:name, :slug, :description, :image, :price, :stock, :category_id) RETURNING id`,
	"deleteProduct": `DELETE FROM product.products WHERE id = :id`,
}

func New(db *sqlx.DB) (*RepoProducts, error) {
	rp := &RepoProducts{
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
func (r *RepoProducts) Validate() error {
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

func (r *RepoProducts) GetProductByName(ctx context.Context, name string) (*productsentity.Product, error) {
	var t productsentity.Product
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getProductByDynamic"]+" WHERE name = :name")

	return &t, stmt.GetContext(ctx, &t, productsentity.Product{Name: name})
}

func (r *RepoProducts) GetProductById(ctx context.Context, id int) (*productsentity.Product, error) {
	var t productsentity.Product
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getProductByDynamic"]+" WHERE id = :id")

	return &t, stmt.GetContext(ctx, &t, productsentity.Product{Id: id})
}

func (r *RepoProducts) Insert(ctx context.Context, payload *productsentity.FormCreateUpdateSchema) int {
	var id int
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertProduct"])
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoProducts) Delete(ctx context.Context, productId int) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["deleteProduct"])
	_, err := stmt.ExecContext(ctx, productsentity.Product{Id: productId})
	if err != nil {
		return err
	}
	return nil
}
