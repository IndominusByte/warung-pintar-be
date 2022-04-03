package categories

import (
	"context"
	"fmt"

	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/categories"
	"github.com/creent-production/cdk-go/pagination"
	"github.com/jmoiron/sqlx"
)

type RepoCategories struct {
	db      *sqlx.DB
	queries map[string]string
	execs   map[string]string
}

var queries = map[string]string{
	"getCategoryByDynamic": `SELECT id, name, created_at, updated_at FROM product.categories`,
}
var execs = map[string]string{
	"insertCategory": `INSERT INTO product.categories (name) VALUES (:name) RETURNING id`,
	"updateCategory": `UPDATE product.categories SET name=:name, updated_at=CURRENT_TIMESTAMP WHERE id = :id`,
	"deleteCategory": `DELETE FROM product.categories WHERE id = :id`,
}

func New(db *sqlx.DB) (*RepoCategories, error) {
	rp := &RepoCategories{
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
func (r *RepoCategories) Validate() error {
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

func (r *RepoCategories) GetCategoryByName(ctx context.Context, name string) (*categoriesentity.Category, error) {
	var t categoriesentity.Category
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getCategoryByDynamic"]+" WHERE name = :name")

	return &t, stmt.GetContext(ctx, &t, categoriesentity.Category{Name: name})
}

func (r *RepoCategories) GetCategoryById(ctx context.Context, id int) (*categoriesentity.Category, error) {
	var t categoriesentity.Category
	stmt, _ := r.db.PrepareNamedContext(ctx, r.queries["getCategoryByDynamic"]+" WHERE id = :id")

	return &t, stmt.GetContext(ctx, &t, categoriesentity.Category{Id: id})
}

func (r *RepoCategories) Insert(ctx context.Context, payload *categoriesentity.JsonCreateUpdateSchema) int {
	var id int
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["insertCategory"])
	stmt.QueryRowxContext(ctx, payload).Scan(&id)

	return id
}

func (r *RepoCategories) GetAllCategoryPaginate(ctx context.Context,
	payload *categoriesentity.QueryParamAllCategorySchema) (*categoriesentity.CategoryPaginate, error) {

	var results categoriesentity.CategoryPaginate

	query := r.queries["getCategoryByDynamic"]
	if len(payload.Q) > 0 {
		query += ` WHERE lower(name) LIKE '%'|| lower(:q) ||'%'`
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

func (r *RepoCategories) Update(ctx context.Context, payload *categoriesentity.JsonCreateUpdateSchema) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["updateCategory"])
	_, err := stmt.ExecContext(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}

func (r *RepoCategories) Delete(ctx context.Context, categoryId int) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["deleteCategory"])
	_, err := stmt.ExecContext(ctx, categoriesentity.Category{Id: categoryId})
	if err != nil {
		return err
	}
	return nil
}
