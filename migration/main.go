package main

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	data := &gsmData{}
	if err := data.loadFromGsm(); err != nil {
		panic(err)
	}

	db, err := sql.Open(data.Driver, data.MasterDsn)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(data.FileMigration, data.Driver, driver)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		panic(err)
	}
}
