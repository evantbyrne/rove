package rove

import (
	"fmt"
	"strings"
)

type ServiceRunCommand struct {
	MachineName string   `arg:"" name:"machine" help:"Name of machine."`
	Name        string   `arg:"" name:"name" help:"Name of service."`
	Image       []string `arg:"" name:"image" passthrough:"" help:"Docker image."`

	ConfigFile string   `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Network    string   `flag:"" name:"network" help:"Network name." default:""`
	Prefix     string   `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
	Publish    []string `flag:"" name:"port" short:"p"`
}

func (cmd *ServiceRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.MachineName, func(conn *SshConnection) error {
			command := fmt.Sprint("docker service create --label rove --name ", cmd.Name)
			if cmd.Network != "" {
				command = fmt.Sprint(command, " --network ", cmd.Prefix, cmd.Network)
			}
			for _, p := range cmd.Publish {
				command = fmt.Sprint(command, " --publish ", p)
			}
			command = fmt.Sprint(command, " ", strings.Join(cmd.Image, " "))
			return conn.
				Run(command, func(res string) error {
					fmt.Printf("✅ Created service '%s'\n", cmd.Name)
					for _, p := range cmd.Publish {
						fmt.Println("\tPublished", p)
					}
					return nil
				}).
				OnError(func(err error) error {
					if err != nil {
						fmt.Println("🚫 Could not add service")
					}
					return err
				}).
				Error
		})
	})
}
