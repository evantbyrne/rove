package rove

import (
	"fmt"

	"github.com/alessio/shellescape"
)

type ServiceDeleteCommand struct {
	Name string `arg:"" name:"name" help:"Name of service."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *ServiceDeleteCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			return conn.
				Run(fmt.Sprint("docker service rm ", shellescape.Quote(cmd.Name)), func(_ string) error {
					fmt.Printf("✅ Deleted service '%s'\n", cmd.Name)
					return nil
				}).
				OnError(func(err error) error {
					if err != nil {
						fmt.Println("🚫 Could not delete service")
					}
					return err
				}).
				Error
		})
	})
}
