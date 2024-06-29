package rove

import (
	"encoding/json"
	"fmt"
	"io"
)

type InspectCommand struct {
	Name string `arg:"" name:"name" help:"Name of service or task."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *InspectCommand) Do(conn SshRunner, stdin io.Reader) error {
	return conn.
		Run(fmt.Sprint("docker service inspect --format json ", cmd.Name), func(res string) error {
			var dockerInspect []map[string]any
			if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
				fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
				return err
			}
			if len(dockerInspect) < 1 {
				return fmt.Errorf("empty docker service inspect JSON: %s", dockerInspect)
			}
			out, err := json.MarshalIndent(dockerInspect[0], "", "    ")
			if err != nil {
				fmt.Println("🚫 Could not format JSON:\n", dockerInspect[0])
				return err
			}
			fmt.Println(string(out))
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("🚫 Could not inspect service")
			}
			return err
		}).
		Error()
}

func (cmd *InspectCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, cmd.Do)
	})
}
