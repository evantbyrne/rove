package rove

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/evantbyrne/trance"
)

type DockerInfoJson struct {
	Swarm DockerInfoSwarmJson `json:"Swarm"`
}

type DockerInfoSwarmJson struct {
	NodeID         string `json:"NodeID"`
	LocalNodeState string `json:"LocalNodeState"`
}

type MachineAddCommand struct {
	Name           string `arg:"" name:"name" help:"Name of remote machine."`
	Address        string `arg:"" name:"address" help:"Public address of remote machine."`
	User           string `arg:"" name:"user" help:"User of remote machine."`
	PrivateKeyFile string `arg:"" name:"pk" help:"Private key file." type:"path"`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Port       int64  `flag:"" name:"port" help:"SSH port of remote machine." default:"22"`
	Prefix     string `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
}

func (cmd *MachineAddCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		key, err := os.ReadFile(cmd.PrivateKeyFile)
		if err != nil {
			return fmt.Errorf("unable to read private key file: %v", err)
		}
		err = SshConnect(fmt.Sprintf("%s:%d", cmd.Address, cmd.Port), cmd.User, key, func(conn *SshConnection) error {
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
						return ErrorSkip{}
					}
					return nil
				}).
				Run(fmt.Sprintf("docker swarm init --advertise-addr %s", cmd.Address), func(_ string) error {
					fmt.Println("✅ Enabled swarm on remote machine")
					return nil
				}).
				OnError(SkipReset).
				Run(fmt.Sprint("docker network ls --format json --filter label=rove --filter name=", cmd.Prefix, "default"), func(res string) error {
					if lines := strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n"); len(lines) > 1 {
						fmt.Println("✅ Default network already exists")
						return ErrorSkip{}
					}
					return nil
				}).
				Run(fmt.Sprint("docker network create --attachable --label rove ", cmd.Prefix, "default"), func(res string) error {
					fmt.Println("✅ Created default network")
					return nil
				}).
				OnError(SkipReset).
				Error
		})
		if err != nil {
			fmt.Println("🚫 Could not add machine")
			return err
		}
		return trance.Query[Machine]().
			Insert(&Machine{
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
			Then(func(_ sql.Result, _ *Machine) error {
				fmt.Printf("✅ Added machine '%s'\n", cmd.Name)
				return nil
			}).
			Error
	})
}
