package rove

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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
	Force      bool   `flag:"" name:"force" help:"Skip confirmations."`
	Port       int64  `flag:"" name:"port" help:"SSH port of remote machine." default:"22"`
	Skip       bool   `flag:"" name:"skip" help:"Skip installation steps on remote machine."`
}

func (cmd *MachineAddCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		exists, err := trance.Query[Machine]().Filter("name", "=", cmd.Name).Exists()
		if err != nil {
			return fmt.Errorf("unable to check if machine exists: %v", err)
		}
		if exists {
			return fmt.Errorf("machine with name '%s' already configured", cmd.Name)
		}

		key, err := os.ReadFile(cmd.PrivateKeyFile)
		if err != nil {
			return fmt.Errorf("unable to read private key file: %v", err)
		}
		err = SshConnect(fmt.Sprintf("%s:%d", cmd.Address, cmd.Port), cmd.User, key, func(conn SshRunner, stdin io.Reader) error {
			fmt.Printf("\nConnected to remote address '%s@%s:%d'.\n", cmd.User, cmd.Address, cmd.Port)

			if cmd.Skip {
				fmt.Println("Skipping install steps.")
				return nil
			}

			missingUfw := false
			mustEnableUfw := true
			mustInstallDocker := true
			mustEnableSwarm := true
			err := conn.
				Run("command -v ufw", func(_ string) error {
					return nil
				}).
				OnError(func(error) error {
					missingUfw = true
					mustEnableUfw = false
					return ErrorSkip{}
				}).
				Run("sudo ufw status", func(res string) error {
					if strings.HasPrefix(res, "Status: active") {
						mustEnableUfw = false
					}
					return nil
				}).
				OnError(SkipReset).
				Run("command -v docker", func(_ string) error {
					mustInstallDocker = false
					return nil
				}).
				OnError(func(error) error {
					return ErrorSkip{}
				}).
				Run("docker info --format json", func(res string) error {
					var dockerInfo DockerInfoJson
					if err := json.Unmarshal([]byte(res), &dockerInfo); err != nil {
						fmt.Println("🚫 Could not parse docker info JSON:\n", res)
						return err
					}
					if dockerInfo.Swarm.NodeID != "" {
						mustEnableSwarm = false
					}
					return nil
				}).
				OnError(SkipReset).
				Error()
			if err != nil {
				return err
			}

			if missingUfw {
				fmt.Println("\n⚠️  Warning: UFW missing. Cannot enable firewall. You should install UFW on the target machine and rerun this command. Alternatively, you may manually disallow access to Docker Swarm management ports.")
			}

			if mustEnableUfw || mustInstallDocker || mustEnableSwarm {
				fmt.Print("\nRove will make the following changes to remote machine:\n\n")
				if mustEnableUfw {
					fmt.Println(" ~ Enable firewall")
				}
				if mustInstallDocker {
					fmt.Println(" ~ Install docker")
				}
				if mustEnableSwarm {
					fmt.Println(" ~ Enable swarm")
				}
				if err := confirmDeployment(cmd.Force, stdin); err != nil {
					return err
				}
				fmt.Println()
			} else {
				fmt.Println("\nNo changes needed.")
			}

			if mustEnableUfw {
				err = conn.
					Run("sudo ufw logging on", func(res string) error {
						fmt.Println("sudo ufw logging on", res)
						return nil
					}).
					Run(fmt.Sprintf("sudo ufw allow %d/tcp", cmd.Port), func(res string) error {
						fmt.Println(fmt.Sprintf("sudo ufw allow %d/tcp", cmd.Port), res)
						return nil
					}).
					Run("sudo ufw --force enable", func(res string) error {
						fmt.Println("sudo ufw enable", res)
						fmt.Println("~ Enabled firewall")
						return nil
					}).
					Error()
				if err != nil {
					return err
				}
			}

			if mustInstallDocker {
				// Via: https://docs.docker.com/engine/install/ubuntu/#install-using-the-convenience-script
				err = conn.
					Run("curl -fsSL https://get.docker.com | sh", func(_ string) error {
						fmt.Println("~ Installed docker")
						return nil
					}).
					Error()
				if err != nil {
					return err
				}
			}

			if mustEnableSwarm {
				err = conn.
					Run(fmt.Sprintf("docker swarm init --advertise-addr %s", cmd.Address), func(_ string) error {
						fmt.Println("~ Enabled swarm")
						return nil
					}).
					Error()
				if err != nil {
					return err
				}
			}
			return nil
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
			Then(func(_ sql.Result, _ *Machine) error {
				return SetPreference(DefaultMachine, cmd.Name).Error
			}).
			Then(func(_ sql.Result, _ *Machine) error {
				fmt.Printf("\nSetup '%s' and set as default machine.\n\n", cmd.Name)
				return nil
			}).
			OnError(func(err error) error {
				fmt.Println("🚫 Could not add machine")
				return err
			}).
			Error
	})
}
