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

type NetworkListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Format     string `flag:"" name:"format" enum:"text,json" help:"Output format. Choices: \"text\", \"json\"." default:"text"`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

type NetworkListJson struct {
	Networks []DockerNetworkLsJson
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
					if cmd.Format == "json" {
						t := NetworkListJson{
							Networks: output,
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
