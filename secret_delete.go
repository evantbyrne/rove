package rove

import (
	"fmt"
	"io"

	"github.com/alessio/shellescape"
)

type SecretDeleteCommand struct {
	Name string `arg:"" name:"name" help:"Name of secret."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Force      bool   `flag:"" name:"force" help:"Skip confirmations."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *SecretDeleteCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn SshRunner, stdin io.Reader) error {
			fmt.Printf("\nRove will delete the '%s' secret.\n", cmd.Name)
			if err := confirmDeployment(cmd.Force, stdin); err != nil {
				return err
			}
			return conn.
				Run(fmt.Sprint("docker secret rm ", shellescape.Quote(cmd.Name)), func(res string) error {
					fmt.Printf("\nDeleted the '%s' secret.\n\n", cmd.Name)
					return nil
				}).
				OnError(func(err error) error {
					if err != nil {
						fmt.Println("🚫 Could not delete secret")
					}
					return err
				}).
				Error()
		})
	})
}
