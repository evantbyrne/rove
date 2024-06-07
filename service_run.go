package rove

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/alessio/shellescape"
)

type DockerServiceInspectJson struct {
	Spec struct {
		TaskTemplate struct {
			ContainerSpec struct {
				Args  []string `json:"Args"`
				Image string   `json:"Image"`
			} `json:"ContainerSpec"`
		} `json:"TaskTemplate"`
		EndpointSpec struct {
			Ports []struct {
				Protocol      string `json:"Protocol"`
				TargetPort    int64  `json:"TargetPort"`
				PublishedPort int64  `json:"PublishedPort"`
				PublishMode   string `json:"PublishMode"`
			} `json:"Ports"`
		} `json:"EndpointSpec"`
		Mode struct {
			Replicated struct {
				Replicas int64 `json:"Replicas"`
			} `json:"Replicated"`
		} `json:"Mode"`
	} `json:"Spec"`
}

type ServiceRunCommand struct {
	Name    string   `arg:"" name:"name" help:"Name of service."`
	Image   string   `arg:"" name:"image" help:"Docker image."`
	Command []string `arg:"" name:"command" optional:"" passthrough:"" help:"Docker command."`

	ConfigFile string   `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Force      bool     `flag:"" name:"force" help:"Skip confirmations."`
	Machine    string   `flag:"" name:"machine" help:"Name of machine." default:""`
	Network    string   `flag:"" name:"network" help:"Network name." default:""`
	Prefix     string   `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
	Publish    []string `flag:"" name:"port" short:"p"`
	Replicas   int64    `flag:"" name:"replicas" default:"1"`
}

func (cmd *ServiceRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			old := &ServiceState{
				Command: make([]string, 0),
				Publish: make([]string, 0),
			}
			new := &ServiceState{
				Command:  cmd.Command,
				Image:    cmd.Image,
				Publish:  cmd.Publish,
				Replicas: fmt.Sprint(cmd.Replicas),
			}
			command := ShellCommand{
				Name: "docker service create",
				Flags: []ShellFlag{
					{
						Check: true,
						Name:  "replicas",
						Value: fmt.Sprintf("%d", cmd.Replicas),
					},
					{
						Check: cmd.Network != "",
						Name:  "network",
						Value: cmd.Prefix + cmd.Network,
					},
				},
				Args: []ShellArg{},
			}
			err := conn.
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
					portsExisting := make([]string, 0)
					for _, entry := range dockerInspect[0].Spec.EndpointSpec.Ports {
						port := fmt.Sprintf("%d:%d", entry.TargetPort, entry.PublishedPort)
						portsExisting = append(portsExisting, port)
						if !slices.Contains(cmd.Publish, port) {
							command.Flags = append(command.Flags, ShellFlag{
								Check: true,
								Name:  "publish-rm",
								Value: port,
							})
						}
					}

					for _, port := range cmd.Publish {
						if !slices.Contains(portsExisting, port) {
							command.Flags = append(command.Flags, ShellFlag{
								Check: true,
								Name:  "publish-add",
								Value: port,
							})
						}
					}
					old.Command = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Args
					old.Image = strings.Split(dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Image, "@")[0]
					old.Publish = portsExisting
					old.Replicas = fmt.Sprint(dockerInspect[0].Spec.Mode.Replicated.Replicas)

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
						Value: shellescape.Quote(cmd.Name),
					})
					return nil
				}).
				OnError(SkipReset).
				Error
			if err != nil {
				fmt.Println("🚫 Could not create deployment plan")
				return err
			}

			diffText, diffStatus := new.Diff(old)
			if command.Name == "docker service create" {
				fmt.Printf("\nRove will create %s:\n\n", cmd.Name)
				fmt.Printf(" + service %s:\n", cmd.Name)
			} else {
				if diffStatus == DiffSame {
					fmt.Printf("\nRove will deploy %s without changes:\n\n", cmd.Name)
					fmt.Printf("   service %s:\n", cmd.Name)
				} else {
					fmt.Printf("\nRove will update %s:\n\n", cmd.Name)
					fmt.Printf(" ~ service %s:\n", cmd.Name)
				}
			}
			fmt.Print(diffText, "\n\n")
			if cmd.Force {
				fmt.Println("Confirmations skipped.")
			} else {
				fmt.Println("Do you want Rove to run this deployment?")
				fmt.Println("  Type 'yes' to approve, or anything else to deny.")
				fmt.Print("  Enter a value: ")
				line, err := bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					fmt.Println("🚫 Could not read from STDIN")
					return err
				}
				if strings.ToLower(strings.TrimSpace(line)) != "yes" {
					return errors.New("🚫 Deployment canceled because response did not match 'yes'")
				}
			}

			fmt.Println("\nDeploying...")

			return conn.
				Run(command.String(), func(res string) error {
					fmt.Printf("\nRove deployed %s.\n\n", cmd.Name)
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
