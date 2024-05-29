package rove

import (
	"errors"
	"fmt"

	"github.com/evantbyrne/trance"
)

type MachineUseCommand struct {
	Name string `arg:"" name:"name" help:"Machine name."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
}

func (cmd *MachineUseCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		err := trance.Query[Machine]().
			Filter("name", "=", cmd.Name).
			First().
			Error
		if err != nil {
			if (errors.Is(err, trance.ErrorNotFound{})) {
				fmt.Printf("🚫 Machine '%s' not configured. Use `rove machine list` to see all configured machines\n", cmd.Name)
			}
			return err
		}
		if err := SetPreference(DefaultMachine, cmd.Name).Error; err != nil {
			fmt.Printf("🚫 Could not set default machine to '%s'\n", cmd.Name)
			return err
		}
		return nil
	})
}
