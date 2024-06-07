package rove

import (
	"fmt"
)

type NetworkDeleteCommand struct {
	Name string `arg:"" name:"name" help:"Name of network."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Prefix     string `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
}

func (cmd *NetworkDeleteCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			return conn.
				Run(fmt.Sprint("docker network rm ", cmd.Prefix, cmd.Name), func(res string) error {
					fmt.Printf("✅ Deleted network '%s'\n", cmd.Name)
					return nil
				}).
				OnError(func(err error) error {
					if err != nil {
						fmt.Println("🚫 Could not delete network")
					}
					return err
				}).
				Error
		})
	})
}
