package carts

import (
	"context"

	// cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"

	"github.com/jmoiron/sqlx"
)

type RepoCarts struct {
	db      *sqlx.DB
	queries map[string]string
	execs   map[string]string
}

var queries = map[string]string{}
var execs = map[string]string{}

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
