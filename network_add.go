package rove

import (
	"fmt"
	"io"

	"github.com/alessio/shellescape"
)

type NetworkAddCommand struct {
	Name string `arg:"" name:"name" help:"Name of network."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Force      bool   `flag:"" name:"force" help:"Skip confirmations."`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *NetworkAddCommand) Do(conn SshRunner, stdin io.Reader) error {
	fmt.Printf("\nRove will create the '%s' network.\n", cmd.Name)
	if err := confirmDeployment(cmd.Force, stdin); err != nil {
		return err
	}
	return conn.
		Run(fmt.Sprint("docker network create --attachable --driver overlay --label rove --scope swarm ", shellescape.Quote(cmd.Name)), func(res string) error {
			fmt.Printf("\nCreated '%s' network.\n\n", cmd.Name)
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("🚫 Could not add network")
			}
			return err
		}).
		Error()
}

func (cmd *NetworkAddCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, cmd.Do)
	})
}
