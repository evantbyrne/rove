package rove

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type DockerVolumeLsJson struct {
	Availability string `json:"Availability"`
	Driver       string `json:"Driver"`
	Group        string `json:"Group"`
	Name         string `json:"Name"`
	Size         string `json:"Size"`
	Status       string `json:"Status"`
}

type VolumeListJson struct {
	Volumes []DockerVolumeLsJson `json:"volumes"`
}

type VolumeJson struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type VolumeListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *VolumeListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, func(conn SshRunner, stdin io.Reader) error {
			return conn.
				Run("docker volume ls --format json --filter label=rove", func(res string) error {
					output := make([]DockerVolumeLsJson, 0)
					for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
						if line != "" {
							var dockerVolumeLs DockerVolumeLsJson
							if err := json.Unmarshal([]byte(line), &dockerVolumeLs); err != nil {
								fmt.Println("🚫 Could not parse docker volume ls JSON:\n", line)
								return err
							}
							output = append(output, dockerVolumeLs)
						}
					}
					if cmd.Json {
						t := VolumeListJson{
							Volumes: output,
						}
						out, err := json.MarshalIndent(t, "", "    ")
						if err != nil {
							fmt.Println("🚫 Could not format JSON:\n", t)
							return err
						}
						fmt.Println(string(out))
					} else {
						for _, dockerVolumeLs := range output {
							fmt.Println(dockerVolumeLs.Name, dockerVolumeLs.Availability)
						}
					}
					return nil
				}).
				Error()
		})
	})
}
