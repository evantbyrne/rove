package rove

import (
	"fmt"

	"github.com/alessio/shellescape"
)

type TaskLogsCommand struct {
	Id string `arg:"" name:"id" help:"ID of task."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Follow     bool   `flag:"" name:"follow" short:"f" help:"Follow log output."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Tail       int64  `flag:"" name:"tail" short:"n" help:"Number of lines to show from the end of the logs."`
	Timeout    string `flag:"" name:"timeout" help:"Timeout duration for tail." default:"1h"`
	Timestamps bool   `flag:"" name:"timestamps" short:"t" help:"Show timestamps."`
}

func (cmd *TaskLogsCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		command := ShellCommand{
			Name: "docker container logs",
			Flags: []ShellFlag{
				{
					Check: cmd.Follow,
					Name:  "follow",
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
					Value: shellescape.Quote(cmd.Id),
				},
			},
		}
		if cmd.Follow && cmd.Timeout != "" {
			command.Name = fmt.Sprintf("timeout --verbose %s %s", cmd.Timeout, command.Name)
		}
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			return conn.Run(command.String(), func(res string) error {
				fmt.Print(res)
				return nil
			}).Error
		})
	})
}
