package rove

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type DockerNetworkLsJson struct {
	Id   string `json:"ID"`
	Name string `json:"Name"`
}

type NetworkListJson struct {
	Networks []NetworkJson `json:"networks"`
}

type NetworkJson struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type NetworkListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *NetworkListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn SshRunner, stdin io.Reader) error {
			return conn.
				Run("docker network ls --format json --filter label=rove", func(res string) error {
					output := make([]DockerNetworkLsJson, 0)
					for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
						if line != "" {
							var dockerNetworkLs DockerNetworkLsJson
							if err := json.Unmarshal([]byte(line), &dockerNetworkLs); err != nil {
								fmt.Println("🚫 Could not parse docker network ls JSON:\n", line)
								return err
							}
							output = append(output, dockerNetworkLs)
						}
					}
					if cmd.Json {
						var t NetworkListJson
						for _, network := range output {
							t.Networks = append(t.Networks, NetworkJson{
								Id:   network.Id,
								Name: network.Name,
							})
						}
						out, err := json.MarshalIndent(t, "", "    ")
						if err != nil {
							fmt.Println("🚫 Could not format JSON:\n", t)
							return err
						}
						fmt.Println(string(out))
					} else {
						for _, dockerNetworkLs := range output {
							fmt.Println(dockerNetworkLs.Id, dockerNetworkLs.Name)
						}
					}
					return nil
				}).
				Error()
		})
	})
}
