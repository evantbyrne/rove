package main

import (
	"github.com/alecthomas/kong"
	"github.com/evantbyrne/rove"
	"github.com/evantbyrne/trance"
	"github.com/evantbyrne/trance/sqlitedialect"

	_ "modernc.org/sqlite"
)

var cli struct {
	Machine struct {
		Add  rove.MachineAddCommand  `cmd:""`
		List rove.MachineListCommand `cmd:""`
	} `cmd:"" help:"Manage machines."`
	Network struct {
		Add    rove.NetworkAddCommand    `cmd:""`
		Delete rove.NetworkDeleteCommand `cmd:""`
		List   rove.NetworkListCommand   `cmd:""`
	} `cmd:"" help:"Manage networks."`
	Run rove.RunCommand `cmd:"" help:"Run containers"`
}

func main() {
	trance.SetDialect(sqlitedialect.SqliteDialect{})
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
