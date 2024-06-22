package rove

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TaskListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

type TaskListJson struct {
	Tasks []TaskListEntryJson `json:"tasks"`
}

type TaskListEntryJson struct {
	Id       string                `json:"id"`
	Image    string                `json:"image"`
	Command  []string              `json:"command"`
	Ports    []ServiceListPortJson `json:"ports"`
	Replicas string                `json:"replicas"`
}

func (cmd *TaskListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn SshRunner) error {
			output := TaskListJson{
				Tasks: make([]TaskListEntryJson, 0),
			}

			if err := conn.Run("docker service ls --format json --filter label=rove=task", func(res string) error {
				for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
					if line != "" {
						var dockerServiceLs DockerServiceLsJson
						if err := json.Unmarshal([]byte(line), &dockerServiceLs); err != nil {
							fmt.Println("🚫 Could not parse docker service ls JSON:\n", line)
							return err
						}
						output.Tasks = append(output.Tasks, TaskListEntryJson{
							Id:       dockerServiceLs.Id,
							Image:    dockerServiceLs.Image,
							Ports:    make([]ServiceListPortJson, 0),
							Replicas: dockerServiceLs.Replicas,
						})
					}
				}
				return nil
			}).Error(); err != nil {
				return err
			}

			for i, task := range output.Tasks {
				if err := conn.Run(fmt.Sprint("docker service inspect ", task.Id), func(res string) error {
					var dockerInspect []DockerServiceInspectJson
					if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
						fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
						return err
					}
					output.Tasks[i].Command = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Args
					for _, entry := range dockerInspect[0].Spec.EndpointSpec.Ports {
						output.Tasks[i].Ports = append(output.Tasks[i].Ports, ServiceListPortJson(entry))
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
				for _, task := range output.Tasks {
					ports := []string{}
					for _, entry := range task.Ports {
						ports = append(ports, fmt.Sprintf("%d:%d/%s", entry.TargetPort, entry.PublishedPort, entry.Protocol))
					}
					fmt.Println(task.Id, task.Image, task.Command, task.Replicas, strings.Join(ports, ","))
				}
			}
			return nil
		})
	})
}
