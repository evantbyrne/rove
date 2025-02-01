package rove

import (
	"fmt"
	"io"
	"os"

	"github.com/alessio/shellescape"
	"github.com/pkg/sftp"
)

type LoginCommand struct {
	Username     string   `arg:"" name:"username" help:"Docker registery username."`
	PasswordFile *os.File `arg:"" name:"password-file" help:"Password/token file. Use dash (-) to read from STDIN."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Registry   string `flag:"" name:"registry" help:"Docker registry server."`
}

func (cmd *LoginCommand) Do(conn SshRunner, stdin io.Reader) error {
	fileName := cmd.Username + ".txt"
	secret, err := io.ReadAll(cmd.PasswordFile)
	if err != nil {
		return err
	}

	if connReal, ok := conn.(*SshConnection); ok {
		transfer, err := sftp.NewClient(connReal.Client)
		if err != nil {
			return err
		}
		defer transfer.Close()

		fh, err := transfer.Create(fileName)
		if err != nil {
			return err
		}
		if _, err := fh.Write([]byte(secret)); err != nil {
			return err
		}
		fh.Close()
	}

	command := fmt.Sprintf("cat %s | docker login --username %s --password-stdin", fileName, shellescape.Quote(cmd.Username))
	registryName := "docker.io"
	if cmd.Registry != "" {
		command = fmt.Sprintf("%s %s", command, shellescape.Quote(cmd.Registry))
		registryName = cmd.Registry
	}

	err = conn.
		Run(command, func(_ string) error {
			fmt.Printf("\nRemote machine logged into '%s' with username '%s'.\n\n", registryName, cmd.Username)
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("\n🚫 Could not login to registry")
			}
			return err
		}).
		Error()

	conn.Run(fmt.Sprintf("rm %s", fileName), func(_ string) error {
		return nil
	})

	return err
}

func (cmd *LoginCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, cmd.Do)
	})
}
