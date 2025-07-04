package rove

import (
	"fmt"
	"io"

	"github.com/alessio/shellescape"
)

type TaskRunCommand struct {
	Image   string   `arg:"" name:"image" help:"Docker image."`
	Command []string `arg:"" name:"command" optional:"" passthrough:"" help:"Docker command."`

	ConfigFile string   `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Env        []string `flag:"" name:"env" short:"e" sep:"none"`
	Force      bool     `flag:"" name:"force" help:"Skip confirmations."`
	Init       bool     `flag:"" name:"init"`
	Local      bool     `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string   `flag:"" name:"machine" help:"Name of machine." default:""`
	Mounts     []string `flag:"" name:"mount" sep:"none"`
	Networks   []string `flag:"" name:"network" help:"Network name."`
	Publish    []string `flag:"" name:"publish" short:"p" sep:"none"`
	Replicas   int64    `flag:"" name:"replicas" default:"1"`
	Secrets    []string `flag:"" name:"secret" sep:"none"`
	User       string   `flag:"" name:"user" short:"u"`
	Verbose    bool     `flag:"" name:"verbose"`
	WorkDir    string   `flag:"" name:"workdir" short:"w"`
}

func (cmd *TaskRunCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, func(conn SshRunner, stdin io.Reader) error {
			old := &ServiceState{}
			new := &ServiceState{
				Command:  cmd.Command,
				Env:      cmd.Env,
				Image:    cmd.Image,
				Init:     cmd.Init,
				Mounts:   cmd.Mounts,
				Networks: cmd.Networks,
				Publish:  cmd.Publish,
				Replicas: fmt.Sprint(cmd.Replicas),
				Secrets:  cmd.Secrets,
				User:     cmd.User,
				WorkDir:  cmd.WorkDir,
			}
			command := ShellCommand{
				Name: "docker service create --detach --no-healthcheck --quiet",
				Flags: []ShellFlag{
					{
						Check: true,
						Name:  "label",
						Value: "rove=task",
					},
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
						Name:  "restart-condition",
						Value: "none",
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
				Args: []ShellArg{
					{
						Check: true,
						Value: shellescape.Quote(cmd.Image),
					},
				},
			}
			for _, arg := range cmd.Command {
				command.Args = append(command.Args, ShellArg{
					Check: true,
					Value: shellescape.Quote(arg),
				})
			}
			for _, env := range cmd.Env {
				command.Flags = append(command.Flags, ShellFlag{
					Check: env != "",
					Name:  "env",
					Value: env,
				})
			}
			for _, mount := range cmd.Mounts {
				command.Flags = append(command.Flags, ShellFlag{
					Check: mount != "",
					Name:  "mount",
					Value: mount,
				})
			}
			for _, network := range cmd.Networks {
				command.Flags = append(command.Flags, ShellFlag{
					Check: network != "",
					Name:  "network",
					Value: network,
				})
			}
			for _, port := range cmd.Publish {
				command.Flags = append(command.Flags, ShellFlag{
					Check: port != "",
					Name:  "publish",
					Value: port,
				})
			}
			for _, secret := range cmd.Secrets {
				command.Flags = append(command.Flags, ShellFlag{
					Check: secret != "",
					Name:  "secret",
					Value: secret,
				})
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

			diffText, _ := new.Diff(old)
			fmt.Print("\nRove will deploy:\n\n")
			fmt.Println(" + task:")
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
					fmt.Print("\nRove deployed task: ", res, "\n")
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
