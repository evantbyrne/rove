package main

import (
	"github.com/alecthomas/kong"
	"github.com/evantbyrne/rove"
	"github.com/evantbyrne/trance"
	"github.com/evantbyrne/trance/sqlitedialect"
)

var cli struct {
	Inspect rove.InspectCommand `cmd:"" help:"Inspect services and tasks."`
	Login   rove.LoginCommand   `cmd:"" help:"Log into docker registries."`
	Logout  rove.LogoutCommand  `cmd:"" help:"Log out of docker registries."`
	Logs    rove.LogsCommand    `cmd:"" help:"View logs."`
	Machine struct {
		Add    rove.MachineAddCommand    `cmd:""`
		Delete rove.MachineDeleteCommand `cmd:""`
		List   rove.MachineListCommand   `cmd:""`
		Use    rove.MachineUseCommand    `cmd:""`
	} `cmd:"" help:"Manage machines."`
	Network struct {
		Add    rove.NetworkAddCommand    `cmd:""`
		Delete rove.NetworkDeleteCommand `cmd:""`
		List   rove.NetworkListCommand   `cmd:""`
	} `cmd:"" help:"Manage networks."`
	Secret struct {
		Create rove.SecretCreateCommand `cmd:""`
		Delete rove.SecretDeleteCommand `cmd:""`
		List   rove.SecretListCommand   `cmd:""`
	} `cmd:"" help:"Manage secrets."`
	Service struct {
		Delete rove.ServiceDeleteCommand `cmd:""`
		List   rove.ServiceListCommand   `cmd:""`
		Run    rove.ServiceRunCommand    `cmd:""`
	} `cmd:"" help:"Manage services."`
	Task struct {
		List rove.TaskListCommand `cmd:""`
		Run  rove.TaskRunCommand  `cmd:""`
	} `cmd:"" help:"Manage one-off tasks."`
}

func main() {
	trance.SetDialect(sqlitedialect.SqliteDialect{})
	ctx := kong.Parse(&cli, kong.UsageOnError())
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
