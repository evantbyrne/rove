package rove

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alessio/shellescape"
)

type InspectCommand struct {
	Name string `arg:"" name:"name" help:"Name of service or task."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Json       bool   `flag:"" name:"json"`
}

func (cmd *InspectCommand) Do(conn SshRunner, stdin io.Reader) error {
	return conn.
		Run(fmt.Sprint("docker service inspect --format json ", cmd.Name), func(res string) error {
			if cmd.Json {
				var dockerInspect []map[string]any
				if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
					fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
					return err
				}
				if len(dockerInspect) < 1 {
					return fmt.Errorf("empty docker service inspect JSON: %s", dockerInspect)
				}
				out, err := json.MarshalIndent(dockerInspect[0], "", "    ")
				if err != nil {
					fmt.Println("🚫 Could not format JSON:\n", dockerInspect[0])
					return err
				}
				fmt.Println(string(out))
			} else {
				var dockerInspect []DockerServiceInspectJson
				if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
					fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
					return err
				}
				old := &ServiceState{}
				old.Env = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Env
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
				}

				// Ports
				portsExisting := make([]string, 0)
				for _, entry := range dockerInspect[0].Spec.EndpointSpec.Ports {
					port := fmt.Sprintf("%d:%d", entry.TargetPort, entry.PublishedPort)
					portsExisting = append(portsExisting, port)
				}

				// Secrets
				secretsExisting := make([]string, 0)
				for _, secret := range dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Secrets {
					secretsExisting = append(secretsExisting, secret.SecretName)
				}

				old.Command = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Args
				old.Image = strings.Split(dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Image, "@")[0]
				old.Init = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Init
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

				diffText, _ := old.Diff(old)
				fmt.Printf("\nCurrent state of %s:\n\n", cmd.Name)
				fmt.Printf("   service %s:\n", cmd.Name)
				fmt.Print(diffText, "\n\n")
			}
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("🚫 Could not inspect service")
			}
			return err
		}).
		Error()
}

func (cmd *InspectCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, cmd.Do)
	})
}
