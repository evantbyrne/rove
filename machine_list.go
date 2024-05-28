package rove

import (
	"encoding/json"
	"fmt"

	"github.com/evantbyrne/trance"
)

type MachineListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Format     string `flag:"" name:"format" enum:"text,json" help:"Output format. Choices: \"text\", \"json\"." default:"text"`
}

type MachineListJson struct {
	Machines []*Machine `json:"machines"`
}

func (cmd *MachineListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return trance.Query[Machine]().
			Sort("name").
			All().
			Then(func(machines []*Machine) error {
				if cmd.Format == "json" {
					to := MachineListJson{
						Machines: machines,
					}
					out, err := json.MarshalIndent(to, "", "    ")
					if err != nil {
						fmt.Println("🚫 Could not format JSON:\n", to)
						return err
					}
					fmt.Println(string(out))
				} else {
					for _, machine := range machines {
						fmt.Println(machine.Name, machine.Address)
					}
				}
				return nil
			}).
			Error
	})
}
