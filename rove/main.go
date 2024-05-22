package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/evantbyrne/rove"
	"github.com/evantbyrne/rove/migrations"
	"github.com/evantbyrne/trance"
	"github.com/evantbyrne/trance/sqlitedialect"

	_ "modernc.org/sqlite"
)

type MachineCommand struct {
	Add MachineCommandAdd `cmd:""`
}

type MachineCommandAdd struct {
	Name           string `arg:"" name:"name" help:"Name of remote machine."`
	Address        string `arg:"" name:"address" help:"Public address of remote machine."`
	User           string `arg:"" name:"user" help:"User of remote machine."`
	PrivateKeyFile string `arg:"" name:"pk" help:"Private key file." type:"path"`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
}

func (cmd *MachineCommandAdd) Run() error {
	trance.SetDialect(sqlitedialect.SqliteDialect{})
	db, err := sql.Open("sqlite", fmt.Sprint("file:", cmd.ConfigFile))
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

	key, err := os.ReadFile(cmd.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("unable to read private key file: %v", err)
	}
	err = rove.SshConnect(cmd.Address, cmd.User, key, func(conn *rove.SshConnection) error {
		fmt.Printf("✅ Connected to remote address '%s@%s'\n", cmd.User, cmd.Address)
		return conn.
			Run("sudo docker run hello-world", func(res string) error {
				fmt.Println("✅ Verified remote docker installation")
				err := trance.Query[rove.Machine]().Insert(&rove.Machine{
					Address: cmd.Address,
					KeyPath: cmd.PrivateKeyFile,
					Name:    cmd.Name,
					User:    cmd.User,
				}).Error
				if err == nil {
					fmt.Printf("✅ Added machine '%s'\n", cmd.Name)
				}
				return err
			}).
			Error
	})
	if err != nil {
		fmt.Println("🚫 Could not add machine")
	}
	return err
}

var cli struct {
	Machine MachineCommand `cmd:"" help:"Manage machines."`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
