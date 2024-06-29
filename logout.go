package rove

import (
	"fmt"
	"io"

	"github.com/alessio/shellescape"
)

type LogoutCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Registry   string `flag:"" name:"registry" help:"Docker registry server."`
}

func (cmd *LogoutCommand) Do(conn SshRunner, stdin io.Reader) error {
	command := "docker logout"
	registryName := "docker.io"
	if cmd.Registry != "" {
		command = fmt.Sprintf("%s %s", command, shellescape.Quote(cmd.Registry))
		registryName = cmd.Registry
	}

	return conn.
		Run(command, func(_ string) error {
			fmt.Printf("\nRemote machine logged out of '%s'.\n\n", registryName)
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("\n🚫 Could not login to registry")
			}
			return err
		}).
		Error()
}

func (cmd *LogoutCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, cmd.Do)
	})
}
