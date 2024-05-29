package rove

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/alessio/shellescape"
)

type DockerServiceInspectJson struct {
	Spec struct {
		EndpointSpec struct {
			Ports []struct {
				Protocol      string `json:"Protocol"`
				TargetPort    int64  `json:"TargetPort"`
				PublishedPort int64  `json:"PublishedPort"`
				PublishMode   string `json:"PublishMode"`
			} `json:"Ports"`
		} `json:"EndpointSpec"`
	} `json:"Spec"`
}

type ServiceRunCommand struct {
	Name    string   `arg:"" name:"name" help:"Name of service."`
	Image   string   `arg:"" name:"image" help:"Docker image."`
	Command []string `arg:"" name:"command" optional:"" passthrough:"" help:"Docker command."`

	ConfigFile string   `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Machine    string   `flag:"" name:"machine" help:"Name of machine." default:""`
	Network    string   `flag:"" name:"network" help:"Network name." default:""`
	Prefix     string   `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
	Publish    []string `flag:"" name:"port" short:"p"`
}

func (cmd *ServiceRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			command := ShellCommand{
				Name: "docker service create",
				Flags: []ShellFlag{
					{
						Check: cmd.Network != "",
						Name:  "network",
						Value: cmd.Prefix + cmd.Network,
					},
				},
				Args: []ShellArg{},
			}
			return conn.
				Run(fmt.Sprint("docker service ls --format json --filter label=rove --filter name=", cmd.Name), func(res string) error {
					if lines := strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n"); len(lines) > 1 {
						return nil
					}
					command.Flags = append(command.Flags, ShellFlag{
						Check: true,
						Name:  "name",
						Value: cmd.Name,
					})
					command.Flags = append(command.Flags, ShellFlag{
						Check: true,
						Name:  "label",
						Value: "rove",
					})
					for _, p := range cmd.Publish {
						command.Flags = append(command.Flags, ShellFlag{
							Check: p != "",
							Name:  "publish",
							Value: p,
						})
					}
					command.Args = append(command.Args,
						ShellArg{
							Check: true,
							Value: shellescape.Quote(cmd.Image),
						},
						ShellArg{
							Check: len(cmd.Command) > 0,
							Value: strings.Join(cmd.Command, " "),
						},
					)
					return ErrorSkip{}
				}).
				Run(fmt.Sprint("docker service inspect ", cmd.Name), func(res string) error {
					var dockerInspect []DockerServiceInspectJson
					if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
						fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
						return err
					}
					portsAdd := make([]string, 0)
					portsExisting := make([]string, 0)
					portsRemove := make([]string, 0)
					for _, entry := range dockerInspect[0].Spec.EndpointSpec.Ports {
						port := fmt.Sprintf("%d:%d", entry.TargetPort, entry.PublishedPort)
						portsExisting = append(portsExisting, port)
						if !slices.Contains(cmd.Publish, port) {
							portsRemove = append(portsRemove, port)
							command.Flags = append(command.Flags, ShellFlag{
								Check: true,
								Name:  "publish-rm",
								Value: port,
							})
						}
					}

					for _, port := range cmd.Publish {
						if !slices.Contains(portsExisting, port) {
							portsAdd = append(portsAdd, port)
							command.Flags = append(command.Flags, ShellFlag{
								Check: true,
								Name:  "publish-add",
								Value: port,
							})
						}
					}
					fmt.Println("--- ports:", portsRemove)
					fmt.Println("+++ ports:", portsAdd)

					command.Name = "docker service update"
					command.Flags = append(command.Flags, ShellFlag{
						Check: len(cmd.Command) > 0,
						Name:  "args",
						Value: strings.Join(cmd.Command, " "),
					})
					command.Flags = append(command.Flags, ShellFlag{
						Check: true,
						Name:  "image",
						Value: cmd.Image,
					})
					command.Args = append(command.Args, ShellArg{
						Check: true,
						Value: cmd.Name,
					})
					return nil
				}).
				OnError(SkipReset).
				Run(command.String(), func(res string) error {
					if command.Name == "docker service create" {
						fmt.Printf("✅ Created service '%s'\n", cmd.Name)
					} else {
						fmt.Printf("✅ Updated service '%s'\n", cmd.Name)
					}
					for _, p := range cmd.Publish {
						fmt.Println("\tPublished", p)
					}
					return nil
				}).
				OnError(func(err error) error {
					if err != nil {
						fmt.Println("🚫 Could not deploy service")
					}
					return err
				}).
				Error
		})
	})
}
