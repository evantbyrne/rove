package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/evantbyrne/rove"
	"github.com/evantbyrne/trance"
	"github.com/evantbyrne/trance/sqlitedialect"

	_ "modernc.org/sqlite"
)

type DockerInfoJson struct {
	Swarm DockerInfoSwarmJson `json:"Swarm"`
}

type DockerInfoSwarmJson struct {
	NodeID         string `json:"NodeID"`
	LocalNodeState string `json:"LocalNodeState"`
}

type MachineCommand struct {
	Add MachineCommandAdd `cmd:""`
}

type MachineCommandAdd struct {
	Name           string `arg:"" name:"name" help:"Name of remote machine."`
	Address        string `arg:"" name:"address" help:"Public address of remote machine."`
	User           string `arg:"" name:"user" help:"User of remote machine."`
	PrivateKeyFile string `arg:"" name:"pk" help:"Private key file." type:"path"`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Port       int64  `flag:"" name:"port" help:"SSH port of remote machine." default:"22"`
}

func (cmd *MachineCommandAdd) Run() error {
	return rove.Database(cmd.ConfigFile, func() error {
		key, err := os.ReadFile(cmd.PrivateKeyFile)
		if err != nil {
			return fmt.Errorf("unable to read private key file: %v", err)
		}
		err = rove.SshConnect(fmt.Sprintf("%s:%d", cmd.Address, cmd.Port), cmd.User, key, func(conn *rove.SshConnection) error {
			fmt.Printf("✅ Connected to remote address '%s@%s:%d'\n", cmd.User, cmd.Address, cmd.Port)
			return conn.
				Run("docker info --format json", func(res string) error {
					var dockerInfo DockerInfoJson
					if err := json.Unmarshal([]byte(res), &dockerInfo); err != nil {
						fmt.Println("🚫 Could not parse docker info JSON:\n", res)
						return err
					}
					if dockerInfo.Swarm.NodeID != "" {
						fmt.Println("✅ Remote machine already part of a swarm")
						return rove.ErrorSkip{}
					}
					return nil
				}).
				Run(fmt.Sprintf("docker swarm init --advertise-addr %s", cmd.Address), func(_ string) error {
					fmt.Println("✅ Enabled swarm on remote machine")
					return nil
				}).
				Error
		})
		if err != nil && !errors.Is(err, rove.ErrorSkip{}) {
			fmt.Println("🚫 Could not add machine")
			return err
		}
		return trance.Query[rove.Machine]().
			Insert(&rove.Machine{
				Address: cmd.Address,
				KeyPath: cmd.PrivateKeyFile,
				Name:    cmd.Name,
				Port:    cmd.Port,
				User:    cmd.User,
			}).
			OnError(func(err error) error {
				fmt.Println("🚫 Could not add machine")
				return err
			}).
			Then(func(_ sql.Result, _ *rove.Machine) error {
				fmt.Printf("✅ Added machine '%s'\n", cmd.Name)
				return nil
			}).
			Error
	})
}

var cli struct {
	Machine MachineCommand `cmd:"" help:"Manage machines."`
}

func main() {
	trance.SetDialect(sqlitedialect.SqliteDialect{})
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
