package main

import (
	"fmt"
	"log"
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
		log.Fatalf("unable to read private key file: %v", err)
	}
	client := rove.SshConnect(cmd.Address, cmd.User, key)
	fmt.Println(rove.SshRun(client, "whoami"))
	fmt.Println(rove.SshRun(client, "sudo docker run hello-world"))
	return nil
}

var cli struct {
	Machine MachineCommand `cmd:"" help:"Manage machines."`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
