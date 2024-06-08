package rove

import (
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
)

type RunCommand struct {
	Image   string   `arg:"" name:"image" help:"Docker image."`
	Command []string `arg:"" name:"command" passthrough:"" help:"Docker command." optional:""`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Follow     bool   `flag:"" name:"follow" short:"f" help:"Wait for container, attach stderr and stdout streams."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Network    string `flag:"" name:"network" help:"Network name." default:"rove"`
}

func (cmd *RunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			command := ShellCommand{
				Name: "docker run --label rove",
				Flags: []ShellFlag{
					{
						Check: cmd.Network != "",
						Name:  "network",
						Value: cmd.Network,
					},
					{
						Check: !cmd.Follow,
						Name:  "detach",
						Value: "",
					},
				},
				Args: []ShellArg{
					{
						Check: true,
						Value: shellescape.Quote(cmd.Image),
					},
					{
						Check: len(cmd.Command) > 0,
						Value: strings.Join(cmd.Command, " "),
					},
				},
			}
			return conn.
				Run(command.String(), func(res string) error {
					fmt.Print(res)
					return nil
				}).
				Error
		})
	})
}
