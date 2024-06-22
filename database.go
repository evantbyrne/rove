package rove

import (
	"database/sql"
	"fmt"

	"github.com/evantbyrne/rove/migrations"
	"github.com/evantbyrne/trance"
	"github.com/evantbyrne/trance/sqlitedialect"

	_ "modernc.org/sqlite"
)

func Database(file string, callback func() error) error {
	trance.SetDialect(sqlitedialect.SqliteDialect{})
	db, err := sql.Open("sqlite", fmt.Sprint("file:", file))
	if err != nil {
		return err
	}
	defer db.Close()

	trance.UseDatabase(db)

	_, err = trance.MigrateUp([]trance.Migration{
		migrations.Migration0001Init{},
	})
	if err != nil {
		return err
	}

	return callback()
}

func testDatabase(callback func() error) error {
	trance.SetDialect(sqlitedialect.SqliteDialect{})
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return err
	}
	defer db.Close()

	trance.UseDatabase(db)

	_, err = trance.MigrateUp([]trance.Migration{
		migrations.Migration0001Init{},
	})
	if err != nil {
		return err
	}

	return callback()
}
