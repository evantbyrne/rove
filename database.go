package rove

import (
	"database/sql"
	"fmt"

	"github.com/evantbyrne/rove/migrations"
	"github.com/evantbyrne/trance"
)

func Database(file string, callback func() error) error {
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
