package rove

import (
	"fmt"
	"strings"
)

type ServiceRunCommand struct {
	MachineName string   `arg:"" name:"machine" help:"Name of machine."`
	Name        string   `arg:"" name:"name" help:"Name of service."`
	Image       string   `arg:"" name:"image" help:"Docker image."`
	Command     []string `arg:"" name:"command" optional:"" passthrough:"" help:"Docker command."`

	ConfigFile string   `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Network    string   `flag:"" name:"network" help:"Network name." default:""`
	Prefix     string   `flag:"" name:"prefix" help:"Network prefix." default:"rove."`
	Publish    []string `flag:"" name:"port" short:"p"`
}

func (cmd *ServiceRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.MachineName, func(conn *SshConnection) error {
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
					} else {
						command.Flags = append(command.Flags, ShellFlag{
							Check: true,
							Name:  "name",
							Value: cmd.Name,
						})
						// TODO: Update ports
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
								Value: cmd.Image,
							},
							ShellArg{
								Check: len(cmd.Command) > 0,
								Value: strings.Join(cmd.Command, " "),
							},
						)
					}
					fmt.Println("<<<", command.String())
					return nil
				}).
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
