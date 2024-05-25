package rove

import (
	"errors"
	"fmt"
	"os"

	"github.com/evantbyrne/trance"
)

type NetworkAddCommand struct {
	MachineName string `arg:"" name:"machine" help:"Name of machine."`
	Name        string `arg:"" name:"name" help:"Name of network."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Prefix     string `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
}

func (cmd *NetworkAddCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		machine, err := trance.Query[Machine]().
			Filter("name", "=", cmd.MachineName).
			CollectFirst()
		if err != nil {
			if errors.Is(err, trance.ErrorNotFound{}) {
				return fmt.Errorf("no machine with name '%s' configured", cmd.MachineName)
			}
			return err
		}

		key, err := os.ReadFile(machine.KeyPath)
		if err != nil {
			return fmt.Errorf("unable to read private key file: %v", err)
		}
		err = SshConnect(fmt.Sprintf("%s:%d", machine.Address, machine.Port), machine.User, key, func(conn *SshConnection) error {
			return conn.
				Run(fmt.Sprint("docker network create --attachable --label rove ", cmd.Prefix, cmd.Name), func(res string) error {
					fmt.Printf("✅ Created network '%s'\n", cmd.Name)
					return nil
				}).
				Error
		})
		if err != nil {
			fmt.Println("🚫 Could not add network")
			return err
		}
		return nil
	})
}
