package migrations

import "github.com/evantbyrne/trance"

type machine0001 struct {
	Id      int64  `@:"id" @primary:"true"`
	Address string `@:"address" @length:"255"`
	KeyPath string `@:"key_path" @length:"1024"`
	Name    string `@:"name" @length:"255" @unique:"true"`
	Port    int64  `@:"port"`
	User    string `@:"user" @length:"255"`
}

type preference0001 struct {
	Name  string `@:"name" @length:"255" @primary:"true"`
	Value string `@:"value"  @length:"2048"`
}

type Migration0001Init struct{}

func (m Migration0001Init) Up() error {
	tx, err := trance.Database().Begin()
	if err != nil {
		return err
	}
	if err = trance.Query[machine0001](trance.WeaveConfig{Table: "machine"}).Transaction(tx).TableCreate().Error; err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = trance.Query[preference0001](trance.WeaveConfig{Table: "preference"}).Transaction(tx).TableCreate().Error; err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (m Migration0001Init) Down() error {
	return nil
}
