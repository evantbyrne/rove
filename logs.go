package rove

import (
	"fmt"

	"github.com/alessio/shellescape"
)

type LogsCommand struct {
	Name string `arg:"" name:"name" help:"Name of service."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Follow     bool   `flag:"" name:"follow" short:"f" help:"Follow log output."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Tail       int64  `flag:"" name:"tail" short:"n" help:"Number of lines to show from the end of the logs."`
	Timeout    string `flag:"" name:"timeout" help:"Timeout duration for tail." default:"1h"`
	Timestamps bool   `flag:"" name:"timestamps" short:"t" help:"Show timestamps."`
}

func (cmd *LogsCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		command := ShellCommand{
			Name: "docker service logs",
			Flags: []ShellFlag{
				{
					Check: cmd.Follow,
					Name:  "follow",
					Value: "",
				},
				{
					Check: true,
					Name:  "no-trunc",
					Value: "",
				},
				{
					Check: true,
					Name:  "raw",
					Value: "",
				},
				{
					Check: cmd.Tail > 0,
					Name:  "tail",
					Value: fmt.Sprint(cmd.Tail),
				},
				{
					Check: cmd.Timestamps,
					Name:  "timestamps",
					Value: "",
				},
			},
			Args: []ShellArg{
				{
					Check: true,
					Value: shellescape.Quote(cmd.Name),
				},
			},
		}
		if cmd.Follow && cmd.Timeout != "" {
			command.Name = fmt.Sprintf("timeout --verbose %s %s", cmd.Timeout, command.Name)
		}
		return SshMachineByName(cmd.Machine, func(conn SshRunner) error {
			return conn.Run(command.String(), func(res string) error {
				fmt.Print(res)
				return nil
			}).Error()
		})
	})
}
