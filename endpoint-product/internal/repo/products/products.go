package products

import (
	"context"
	"fmt"

	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/products"
	"github.com/creent-production/cdk-go/pagination"

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

func (r *RepoProducts) GetProductBySlug(ctx context.Context, slug string) (*productsentity.Product, error) {
	var t productsentity.Product
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getProductByDynamic"]+" WHERE slug = :slug")

	return &t, stmt.GetContext(ctx, &t, productsentity.Product{Slug: slug})
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

func (r *RepoProducts) Update(ctx context.Context, payload *productsentity.FormCreateUpdateSchema) error {
	query := `UPDATE product.products SET updated_at=CURRENT_TIMESTAMP`
	if len(payload.Name) > 0 {
		query += `, name=:name`
	}
	if len(payload.Slug) > 0 {
		query += `, slug=:slug`
	}
	if len(payload.Description) > 0 {
		query += `, description=:description`
	}
	if len(payload.Image) > 0 {
		query += `, image=:image`
	}
	if payload.Price > 0 {
		query += `, price=:price`
	}
	if payload.Stock > 0 {
		query += `, stock=:stock`
	}
	if payload.CategoryId > 0 {
		query += `, category_id=:category_id`
	}

	query += ` WHERE id = :id`

	stmt, _ := r.db.PrepareNamedContext(ctx, query)
	_, err := stmt.ExecContext(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoProducts) Delete(ctx context.Context, productId int) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["deleteProduct"])
	_, err := stmt.ExecContext(ctx, productsentity.Product{Id: productId})
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoProducts) GetAllProductPaginate(ctx context.Context,
	payload *productsentity.QueryParamAllProductSchema) (*productsentity.ProductPaginate, error) {

	var results productsentity.ProductPaginate

	query := r.queries["getProductByDynamic"] + ` WHERE 1=1`
	if len(payload.Q) > 0 {
		query += ` AND lower(name) LIKE '%'|| lower(:q) ||'%'`
	}
	if payload.CategoryId > 0 {
		query += ` AND category_id=:category_id`
	}

	if payload.OrderBy == "high_price" {
		query += ` ORDER BY price DESC`
	} else if payload.OrderBy == "low_price" {
		query += ` ORDER BY price ASC`
	} else {
		query += ` ORDER BY id DESC`
	}

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
