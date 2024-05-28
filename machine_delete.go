package rove

import (
	"errors"
	"fmt"

	"github.com/evantbyrne/trance"
)

type MachineDeleteCommand struct {
	Name string `arg:"" name:"machine" help:"Name of machine."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
}

func (cmd *MachineDeleteCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return trance.Query[Machine]().
			Filter("name", "=", cmd.Name).
			First().
			Then(func(_ *Machine) error {
				return trance.Query[Machine]().
					Filter("name", "=", cmd.Name).
					Delete().
					Error
			}).
			Then(func(_ *Machine) error {
				fmt.Printf("✅ Deleted machine '%s'\n", cmd.Name)
				return nil
			}).
			OnError(func(err error) error {
				if errors.Is(err, trance.ErrorNotFound{}) {
					fmt.Printf("✅ Machine '%s' not found\n", cmd.Name)
					return nil
				} else {
					fmt.Printf("🚫 Could not delete machine '%s':\n", cmd.Name)
				}
				return err
			}).
			Error
	})
}
