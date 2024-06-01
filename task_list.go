package rove

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DockerTaskLsJson struct {
	Command      string `json:"Command"`
	CreatedAt    string `json:"CreatedAt"`
	Id           string `json:"ID"`
	Image        string `json:"Image"`
	Labels       string `json:"Labels"`
	LocalVolumes string `json:"LocalVolumes"`
	Mounts       string `json:"Mounts"`
	Names        string `json:"Names"`
	Networks     string `json:"Networks"`
	Ports        string `json:"Ports"`
	RunningFor   string `json:"RunningFor"`
	State        string `json:"State"`
	Status       string `json:"Status"`
}

type TaskListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Last       int64  `flag:"" name:"last" short:"n" help:"Show n last created containers (includes all states)." default:"0"`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

type TaskListJson struct {
	Tasks []TaskListEntryJson `json:"tasks"`
}

type TaskListEntryJson struct {
	Command    string `json:"command"`
	CreatedAt  string `json:"createdAt"`
	Id         string `json:"id"`
	Image      string `json:"image"`
	Names      string `json:"names"`
	RunningFor string `json:"running_for"`
	State      string `json:"state"`
	Status     string `json:"status"`
}

func (cmd *TaskListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			output := TaskListJson{
				Tasks: make([]TaskListEntryJson, 0),
			}

			command := "docker container ls --format json --filter label=rove"
			if cmd.Last > 0 {
				command = fmt.Sprintf("%s --last %d", command, cmd.Last)
			}
			if err := conn.Run(command, func(res string) error {
				for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
					if line != "" {
						var dockerTaskLs DockerTaskLsJson
						if err := json.Unmarshal([]byte(line), &dockerTaskLs); err != nil {
							fmt.Println("🚫 Could not parse docker container ls JSON:\n", line)
							return err
						}
						output.Tasks = append(output.Tasks, TaskListEntryJson{
							Command:    dockerTaskLs.Command,
							CreatedAt:  dockerTaskLs.CreatedAt,
							Id:         dockerTaskLs.Id,
							Image:      dockerTaskLs.Image,
							Names:      dockerTaskLs.Names,
							RunningFor: dockerTaskLs.RunningFor,
							State:      dockerTaskLs.State,
							Status:     dockerTaskLs.Status,
						})
					}
				}
				return nil
			}).Error; err != nil {
				return err
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
					fmt.Println(task.Id, task.Names, task.Image, task.Command, task.CreatedAt, task.RunningFor, task.State, task.Status)
				}
			}
			return nil
		})
	})
}
