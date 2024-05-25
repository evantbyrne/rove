package rove

import (
	"fmt"
	"strings"
)

type RunCommand struct {
	MachineName string   `arg:"" name:"machine" help:"Name of machine."`
	Image       []string `arg:"" name:"image" passthrough:"" help:"Docker image."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Network    string `flag:"" name:"network" help:"Network name." default:""`
	Prefix     string `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
}

func (cmd *RunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.MachineName, func(conn *SshConnection) error {
			command := "docker run --label rove"
			if cmd.Network != "" {
				command = fmt.Sprint(command, " --network ", cmd.Prefix, cmd.Network)
			}
			command = fmt.Sprint(command, " ", strings.Join(cmd.Image, " "))
			return conn.
				Run(command, func(res string) error {
					fmt.Print(res)
					return nil
				}).
				Error
		})
	})
}
