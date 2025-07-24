package rove

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/alessio/shellescape"
)

type VolumeInspectCommand struct {
	Name string `arg:"" name:"name" help:"Name of volume."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *VolumeInspectCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, func(conn SshRunner, stdin io.Reader) error {
			return conn.
				Run(fmt.Sprint("docker volume inspect ", shellescape.Quote(cmd.Name)), func(res string) error {
					var dockerVolumeInspect []map[string]any
					if err := json.Unmarshal([]byte(res), &dockerVolumeInspect); err != nil {
						fmt.Println("🚫 Could not parse docker volume inspect JSON:\n", res)
						return err
					}
					if len(dockerVolumeInspect) < 1 {
						return fmt.Errorf("empty docker volume inspect JSON: %s", dockerVolumeInspect)
					}
					out, err := json.MarshalIndent(dockerVolumeInspect[0], "", "    ")
					if err != nil {
						fmt.Println("🚫 Could not format JSON:\n", dockerVolumeInspect)
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
		})
	})
}
