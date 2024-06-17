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
		TaskTemplate struct {
			ContainerSpec struct {
				Args    []string `json:"Args"`
				Image   string   `json:"Image"`
				Secrets []struct {
					SecretName string `json:"SecretName"`
				} `json:"Secrets"`
			} `json:"ContainerSpec"`
			Networks []struct {
				Target string `json:"Target"`
			} `json:"Networks"`
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
	Networks   []string `flag:"" name:"network" help:"Network name."`
	Publish    []string `flag:"" name:"publish" short:"p" sep:"none"`
	Replicas   int64    `flag:"" name:"replicas" default:"1"`
	Secrets    []string `flag:"" name:"secret" sep:"none"`
}

func (cmd *ServiceRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			old := &ServiceState{}
			new := &ServiceState{
				Command:  cmd.Command,
				Image:    cmd.Image,
				Networks: cmd.Networks,
				Publish:  cmd.Publish,
				Replicas: fmt.Sprint(cmd.Replicas),
				Secrets:  cmd.Secrets,
			}
			command := ShellCommand{
				Name: "docker service create",
				Flags: []ShellFlag{
					{
						Check: true,
						Name:  "replicas",
						Value: fmt.Sprintf("%d", cmd.Replicas),
					},
				},
			}
			err := conn.
				Run(fmt.Sprint("docker service ls --format json --filter label=rove=service --filter name=", cmd.Name), func(res string) error {
					if lines := strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n"); len(lines) > 1 {
						return nil
					}
					command.Flags = append(command.Flags, ShellFlag{
						Check: true,
						Name:  "label",
						Value: "rove=service",
					})
					command.Flags = append(command.Flags, ShellFlag{
						Check: true,
						Name:  "name",
						Value: cmd.Name,
					})
					for _, network := range cmd.Networks {
						command.Flags = append(command.Flags, ShellFlag{
							Check: network != "",
							Name:  "network",
							Value: network,
						})
					}
					for _, p := range cmd.Publish {
						command.Flags = append(command.Flags, ShellFlag{
							Check: p != "",
							Name:  "publish",
							Value: p,
						})
					}
					for _, secret := range cmd.Secrets {
						command.Flags = append(command.Flags, ShellFlag{
							Check: secret != "",
							Name:  "secret",
							Value: secret,
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

					// Networks
					networksExisting := make([]string, 0)
					networksExistingIds := make([]string, 0)
					for _, network := range dockerInspect[0].Spec.TaskTemplate.Networks {
						networksExistingIds = append(networksExistingIds, network.Target)
					}
					if len(networksExistingIds) > 0 {
						commandNetworks := ShellCommand{
							Name: "docker network ls --format json --no-trunc",
						}
						for _, networkId := range networksExistingIds {
							commandNetworks.Flags = append(commandNetworks.Flags, ShellFlag{
								Check: networkId != "",
								Name:  "filter",
								Value: shellescape.Quote("id=" + networkId),
							})
						}
						errNetwork := conn.
							Run(commandNetworks.String(), func(resNetworks string) error {
								for _, line := range strings.Split(strings.ReplaceAll(resNetworks, "\r\n", "\n"), "\n") {
									if line != "" {
										var dockerNetworkLs DockerNetworkLsJson
										if err := json.Unmarshal([]byte(line), &dockerNetworkLs); err != nil {
											fmt.Println("🚫 Could not parse docker network ls JSON:\n", line)
											return err
										}
										networksExisting = append(networksExisting, dockerNetworkLs.Name)
									}
								}
								return nil
							}).
							Error
						if errNetwork != nil {
							return errNetwork
						}
						for _, network := range networksExisting {
							if !slices.Contains(cmd.Networks, network) {
								command.Flags = append(command.Flags, ShellFlag{
									Check: network != "",
									Name:  "network-rm",
									Value: network,
								})
							}
						}
					}
					for _, network := range cmd.Networks {
						if !slices.Contains(networksExisting, network) {
							command.Flags = append(command.Flags, ShellFlag{
								Check: network != "",
								Name:  "network-add",
								Value: network,
							})
						}
					}

					// Ports
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

					// Secrets
					secretsExisting := make([]string, 0)
					for _, secret := range dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Secrets {
						secretsExisting = append(secretsExisting, secret.SecretName)
						if !slices.Contains(cmd.Secrets, secret.SecretName) {
							command.Flags = append(command.Flags, ShellFlag{
								Check: secret.SecretName != "",
								Name:  "secret-rm",
								Value: secret.SecretName,
							})
						}
					}
					for _, secret := range cmd.Secrets {
						if !slices.Contains(secretsExisting, secret) {
							command.Flags = append(command.Flags, ShellFlag{
								Check: secret != "",
								Name:  "secret-add",
								Value: secret,
							})
						}
					}

					old.Command = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Args
					old.Image = strings.Split(dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Image, "@")[0]
					old.Networks = networksExisting
					old.Publish = portsExisting
					old.Replicas = fmt.Sprint(dockerInspect[0].Spec.Mode.Replicated.Replicas)
					old.Secrets = secretsExisting

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
			fmt.Println(diffText)
			if err := confirmDeployment(cmd.Force); err != nil {
				return err
			}

			fmt.Println("\nDeploying...")

			return conn.
				Run(command.String(), func(res string) error {
					fmt.Printf("\nRove deployed '%s'.\n\n", cmd.Name)
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
