package rove

import (
	"fmt"

	"github.com/alessio/shellescape"
)

type NetworkDeleteCommand struct {
	Name string `arg:"" name:"name" help:"Name of network."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Force      bool   `flag:"" name:"force" help:"Skip confirmations."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *NetworkDeleteCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			fmt.Printf("\nRove will delete the '%s' network.\n", cmd.Name)
			if err := confirmDeployment(cmd.Force); err != nil {
				return err
			}
			return conn.
				Run(fmt.Sprint("docker network rm ", shellescape.Quote(cmd.Name)), func(res string) error {
					fmt.Printf("Deleted '%s' network.\n\n", cmd.Name)
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
