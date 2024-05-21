package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/evantbyrne/rove"
)

type MachineCommand struct {
	Add MachineCommandAdd `cmd:""`
}

type MachineCommandAdd struct {
	Address        string `arg:"" name:"address" help:"Public address of remote machine."`
	User           string `arg:"" name:"user" help:"User of remote machine."`
	PrivateKeyFile string `arg:"" name:"pk" help:"Private key file." type:"path"`
}

func (cmd *MachineCommandAdd) Run() error {
	key, err := os.ReadFile(cmd.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("unable to read private key file: %v", err)
	}
	return rove.SshConnect(cmd.Address, cmd.User, key, func(conn *rove.SshConnection) error {
		return conn.
			Run("whoami", func(res string) error {
				fmt.Println(res)
				return nil
			}).
			Run("sudo docker run hello-world", func(res string) error {
				fmt.Println(res)
				return nil
			}).
			Error
	})
}

var cli struct {
	Machine MachineCommand `cmd:"" help:"Manage machines."`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
