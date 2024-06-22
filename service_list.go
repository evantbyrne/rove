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
	Id       string                `json:"id"`
	Image    string                `json:"image"`
	Name     string                `json:"name"`
	Ports    []ServiceListPortJson `json:"ports"`
	Replicas string                `json:"replicas"`
}

type ServiceListPortJson struct {
	Protocol      string `json:"protocol"`
	TargetPort    int64  `json:"target_port"`
	PublishedPort int64  `json:"published_port"`
	PublishMode   string `json:"publish_mode"`
}

func (cmd *ServiceListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn SshRunner) error {
			output := ServiceListJson{
				Services: make([]ServiceListEntryJson, 0),
			}

			if err := conn.Run("docker service ls --format json --filter label=rove=service", func(res string) error {
				for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
					if line != "" {
						var dockerServiceLs DockerServiceLsJson
						if err := json.Unmarshal([]byte(line), &dockerServiceLs); err != nil {
							fmt.Println("🚫 Could not parse docker service ls JSON:\n", line)
							return err
						}
						output.Services = append(output.Services, ServiceListEntryJson{
							Id:       dockerServiceLs.Id,
							Image:    dockerServiceLs.Image,
							Name:     dockerServiceLs.Name,
							Ports:    make([]ServiceListPortJson, 0),
							Replicas: dockerServiceLs.Replicas,
						})
					}
				}
				return nil
			}).Error(); err != nil {
				return err
			}

			for i, service := range output.Services {
				if err := conn.Run(fmt.Sprint("docker service inspect ", service.Name), func(res string) error {
					var dockerInspect []DockerServiceInspectJson
					if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
						fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
						return err
					}
					for _, entry := range dockerInspect[0].Spec.EndpointSpec.Ports {
						output.Services[i].Ports = append(output.Services[i].Ports, ServiceListPortJson(entry))
					}
					return nil
				}).Error(); err != nil {
					return err
				}
			}

			if cmd.Json {
				out, err := json.MarshalIndent(output, "", "    ")
				if err != nil {
					fmt.Println("🚫 Could not format JSON:\n", output)
					return err
				}
				fmt.Println(string(out))
			} else {
				for _, service := range output.Services {
					ports := []string{}
					for _, entry := range service.Ports {
						ports = append(ports, fmt.Sprintf("%d:%d/%s", entry.TargetPort, entry.PublishedPort, entry.Protocol))
					}
					fmt.Println(service.Id, service.Name, service.Image, service.Replicas, strings.Join(ports, ","))
				}
			}
			return nil
		})
	})
}
