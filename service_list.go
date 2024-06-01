package rove

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DockerServiceLsJson struct {
	Id       string `json:"ID"`
	Image    string `json:"Image"`
	Name     string `json:"Name"`
	Replicas string `json:"Replicas"`
}

type ServiceListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

type ServiceListJson struct {
	Services []ServiceListEntryJson `json:"services"`
}

type ServiceListEntryJson struct {
	Id       string `json:"id"`
	Image    string `json:"image"`
	Name     string `json:"name"`
	Replicas string `json:"replicas"`
}

func (cmd *ServiceListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			output := make([]DockerServiceLsJson, 0)

			if err := conn.Run("docker service ls --format json --filter label=rove", func(res string) error {
				for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
					if line != "" {
						var dockerServiceLs DockerServiceLsJson
						if err := json.Unmarshal([]byte(line), &dockerServiceLs); err != nil {
							fmt.Println("🚫 Could not parse docker service ls JSON:\n", line)
							return err
						}
						output = append(output, dockerServiceLs)
					}
				}
				return nil
			}).Error; err != nil {
				return err
			}

			if cmd.Json {
				t := ServiceListJson{
					Services: make([]ServiceListEntryJson, 0),
				}
				for _, dockerServiceLs := range output {
					t.Services = append(t.Services, ServiceListEntryJson(dockerServiceLs))
				}
				out, err := json.MarshalIndent(t, "", "    ")
				if err != nil {
					fmt.Println("🚫 Could not format JSON:\n", t)
					return err
				}
				fmt.Println(string(out))
			} else {
				for _, dockerServiceLs := range output {
					fmt.Println(dockerServiceLs.Id, dockerServiceLs.Name, dockerServiceLs.Image, dockerServiceLs.Replicas)
				}
			}

			return nil
		})
	})
}
