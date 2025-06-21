package rove

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/alessio/shellescape"
)

type DockerServiceInspectJson struct {
	Spec struct {
		TaskTemplate struct {
			ContainerSpec struct {
				Args    []string `json:"Args"`
				Dir     string   `json:"Dir"`
				Env     []string `json:"Env"`
				Image   string   `json:"Image"`
				Init    bool     `json:"Init"`
				Secrets []struct {
					SecretName string `json:"SecretName"`
				} `json:"Secrets"`
				User string `json:"User"`
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
		UpdateConfig struct {
			Delay         uint64 `json:"Delay"`
			FailureAction string `json:"FailureAction"`
			Order         string `json:"Order"`
			Parallelism   int64  `json:"Parallelism"`
		} `json:"UpdateConfig"`
	} `json:"Spec"`
}

type ServiceRunCommand struct {
	Name    string   `arg:"" name:"name" help:"Name of service."`
	Image   string   `arg:"" name:"image" help:"Docker image."`
	Command []string `arg:"" name:"command" optional:"" passthrough:"" help:"Docker command."`

	ConfigFile          string   `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Env                 []string `flag:"" name:"env" short:"e" sep:"none"`
	Force               bool     `flag:"" name:"force" help:"Skip confirmations."`
	Init                bool     `flag:"" name:"init"`
	Local               bool     `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine             string   `flag:"" name:"machine" help:"Name of machine." default:""`
	Networks            []string `flag:"" name:"network" help:"Network name."`
	Publish             []string `flag:"" name:"publish" short:"p" sep:"none"`
	Replicas            int64    `flag:"" name:"replicas" default:"1"`
	Secrets             []string `flag:"" name:"secret" sep:"none"`
	UpdateDelay         string   `flag:"" name:"update-delay"`
	UpdateFailureAction string   `flag:"" name:"update-failure-action"`
	UpdateOrder         string   `flag:"" name:"update-order"`
	UpdateParallelism   int64    `flag:"" name:"update-parallelism" default:"1"`
	User                string   `flag:"" name:"user" short:"u"`
	Verbose             bool     `flag:"" name:"verbose"`
	WorkDir             string   `flag:"" name:"workdir" short:"w"`
}

func (cmd *ServiceRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, func(conn SshRunner, stdin io.Reader) error {
			old := &ServiceState{}
			new := &ServiceState{
				Command:             cmd.Command,
				Env:                 cmd.Env,
				Image:               cmd.Image,
				Init:                cmd.Init,
				Networks:            cmd.Networks,
				Publish:             cmd.Publish,
				Replicas:            fmt.Sprint(cmd.Replicas),
				Secrets:             cmd.Secrets,
				UpdateDelay:         cmd.UpdateDelay,
				UpdateOrder:         cmd.UpdateOrder,
				UpdateFailureAction: cmd.UpdateFailureAction,
				UpdateParallelism:   fmt.Sprint(cmd.UpdateParallelism),
				User:                cmd.User,
				WorkDir:             cmd.WorkDir,
			}

			updateDelay := ternary(cmd.UpdateDelay == "", "0s", cmd.UpdateDelay)
			updateFailureAction := ternary(cmd.UpdateFailureAction == "", "pause", cmd.UpdateFailureAction)
			updateOrder := ternary(cmd.UpdateOrder == "", "stop-first", cmd.UpdateOrder)
			if new.UpdateParallelism == "1" {
				new.UpdateParallelism = ""
			}

			command := ShellCommand{
				Name: "docker service create",
				Flags: []ShellFlag{
					{
						Check: cmd.Init,
						Name:  "init",
					},
					{
						Check: true,
						Name:  "replicas",
						Value: fmt.Sprintf("%d", cmd.Replicas),
					},
					{
						Check: true,
						Name:  "update-delay",
						Value: updateDelay,
					},
					{
						Check: true,
						Name:  "update-failure-action",
						Value: updateFailureAction,
					},
					{
						Check: true,
						Name:  "update-order",
						Value: updateOrder,
					},
					{
						Check: true,
						Name:  "update-parallelism",
						Value: fmt.Sprint(cmd.UpdateParallelism),
					},
					{
						AllowEmpty: true,
						Check:      true,
						Name:       "user",
						Value:      cmd.User,
					},
					{
						AllowEmpty: true,
						Check:      true,
						Name:       "workdir",
						Value:      cmd.WorkDir,
					},
				},
			}

			commandPull := ShellCommand{
				Name: "docker image pull",
				Args: []ShellArg{
					{
						Check: true,
						Value: shellescape.Quote(cmd.Image),
					},
				},
				Flags: []ShellFlag{
					{
						Check: true,
						Name:  "quiet",
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
					for _, env := range cmd.Env {
						command.Flags = append(command.Flags, ShellFlag{
							Check: env != "",
							Name:  "env",
							Value: env,
						})
					}
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
					)
					for _, arg := range cmd.Command {
						command.Args = append(command.Args, ShellArg{
							Check: true,
							Value: shellescape.Quote(arg),
						})
					}
					return ErrorSkip{}
				}).
				Run(fmt.Sprint("docker service inspect ", cmd.Name), func(res string) error {
					var dockerInspect []DockerServiceInspectJson
					if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
						fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
						return err
					}

					// Environment variables
					old.Env = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Env
					for _, env := range old.Env {
						if !slices.Contains(cmd.Env, env) {
							envName := strings.Split(env, "=")[0]
							command.Flags = append(command.Flags, ShellFlag{
								Check: envName != "",
								Name:  "env-rm",
								Value: envName,
							})
						}
					}
					for _, env := range new.Env {
						if !slices.Contains(old.Env, env) {
							command.Flags = append(command.Flags, ShellFlag{
								Check: env != "",
								Name:  "env-add",
								Value: env,
							})
						}
					}

					// TODO: Mounts

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
							Error()
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
					old.Init = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Init
					if old.Init {
						new.Init = true
					}
					old.Networks = networksExisting
					old.Publish = portsExisting
					old.Replicas = fmt.Sprint(dockerInspect[0].Spec.Mode.Replicated.Replicas)
					old.Secrets = secretsExisting
					if dockerInspect[0].Spec.UpdateConfig.Delay != 0 {
						delayNs, _ := time.ParseDuration(fmt.Sprint(dockerInspect[0].Spec.UpdateConfig.Delay, "ns"))
						old.UpdateDelay, _ = strings.CutSuffix(delayNs.String(), "m0s")
					}
					old.UpdateFailureAction = dockerInspect[0].Spec.UpdateConfig.FailureAction
					if old.UpdateFailureAction == "pause" {
						old.UpdateFailureAction = ""
					}
					old.UpdateOrder = dockerInspect[0].Spec.UpdateConfig.Order
					if old.UpdateOrder == "stop-first" {
						old.UpdateOrder = ""
					}
					old.UpdateParallelism = fmt.Sprint(dockerInspect[0].Spec.UpdateConfig.Parallelism)
					if old.UpdateParallelism == "1" {
						old.UpdateParallelism = ""
					}
					old.User = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.User
					old.WorkDir = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Dir

					command.Name = "docker service update"
					command.Flags = append(command.Flags, ShellFlag{
						Check: len(cmd.Command) > 0,
						Name:  "args",
						Value: shellescape.QuoteCommand(cmd.Command),
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
				Error()
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
			if err := confirmDeployment(cmd.Force, stdin); err != nil {
				return err
			}

			fmt.Println("\nDeploying...")

			return conn.
				Run(commandPull.String(), func(res string) error {
					if cmd.Verbose {
						fmt.Printf("\n[verbose] %s: %s", commandPull.String(), res)
					}
					return nil
				}).
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
				Error()
		})
	})
}
