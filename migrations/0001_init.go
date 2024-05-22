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
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (m Migration0001Init) Down() error {
	if err := trance.Query[machine0001](trance.WeaveConfig{Table: "machine"}).TableDrop(trance.TableDropConfig{IfExists: true}).Error; err != nil {
		return err
	}
	return nil
}
