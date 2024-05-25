package rove

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/evantbyrne/trance"
)

type RunCommand struct {
	MachineName string   `arg:"" name:"machine" help:"Name of machine."`
	Image       []string `arg:"" name:"image" passthrough:"" help:"Docker image."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
}

func (cmd *RunCommand) Run() error {
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
		return SshConnect(fmt.Sprintf("%s:%d", machine.Address, machine.Port), machine.User, key, func(conn *SshConnection) error {
			return conn.
				Run(fmt.Sprintf("docker run %s", strings.Join(cmd.Image, " ")), func(res string) error {
					fmt.Print(res)
					return nil
				}).
				Error
		})
	})
}
