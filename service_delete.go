package rove

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"
)

type ServiceDeleteCommand struct {
	Name string `arg:"" name:"name" help:"Name of service."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Force      bool   `flag:"" name:"force" help:"Skip confirmations."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *ServiceDeleteCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn *SshConnection) error {
			old := &ServiceState{
				Command: make([]string, 0),
				Publish: make([]string, 0),
			}
			err := conn.
				Run(fmt.Sprint("docker service inspect ", cmd.Name), func(res string) error {
					var dockerInspect []DockerServiceInspectJson
					if err := json.Unmarshal([]byte(res), &dockerInspect); err != nil {
						fmt.Println("🚫 Could not parse docker service inspect JSON:\n", res)
						return err
					}
					old.Command = dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Args
					old.Image = strings.Split(dockerInspect[0].Spec.TaskTemplate.ContainerSpec.Image, "@")[0]
					for _, entry := range dockerInspect[0].Spec.EndpointSpec.Ports {
						port := fmt.Sprintf("%d:%d", entry.TargetPort, entry.PublishedPort)
						old.Publish = append(old.Publish, port)
					}
					old.Replicas = fmt.Sprint(dockerInspect[0].Spec.Mode.Replicated.Replicas)
					return nil
				}).
				Error
			if err != nil {
				fmt.Println("🚫 Could not create deployment plan")
				return err
			}

			diffText, _ := (&ServiceState{}).Diff(old)
			fmt.Printf("\nRove will delete %s:\n\n", cmd.Name)
			fmt.Printf(" - service %s:\n", cmd.Name)
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
				Run(fmt.Sprint("docker service rm ", shellescape.Quote(cmd.Name)), func(_ string) error {
					fmt.Printf("\nRove deleted '%s'.\n\n", cmd.Name)
					return nil
				}).
				OnError(func(err error) error {
					if err != nil {
						fmt.Println("🚫 Could not delete service")
					}
					return err
				}).
				Error
		})
	})
}
