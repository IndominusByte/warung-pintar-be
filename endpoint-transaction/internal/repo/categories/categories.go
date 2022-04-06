package categories

import (
	"context"

	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/categories"
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

func (r *RepoCategories) Delete(ctx context.Context, categoryId int) error {
	stmt, _ := r.db.PrepareNamedContext(ctx, r.execs["deleteCategory"])
	_, err := stmt.ExecContext(ctx, categoriesentity.Category{Id: categoryId})
	if err != nil {
		return err
	}
	return nil
}
