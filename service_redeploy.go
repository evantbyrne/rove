package rove

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alessio/shellescape"
)

type ServiceRedeployCommand struct {
	Name string `arg:"" name:"name" help:"Name of service or task."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
	Verbose    bool   `flag:"" name:"verbose"`
}

func (cmd *ServiceRedeployCommand) Do(conn SshRunner, stdin io.Reader) error {
	var dockerServiceLs DockerServiceLsJson
	commandList := fmt.Sprint("docker service ls --format json --filter label=rove=service --filter name=", shellescape.Quote(cmd.Name))
	commandUpdate := fmt.Sprint("docker service update --force ", shellescape.Quote(cmd.Name))
	commandPull := ShellCommand{
		Name: "docker image pull",
		Flags: []ShellFlag{
			{
				Check: true,
				Name:  "quiet",
			},
		},
	}

	return conn.
		Run(commandList, func(res string) error {
			if cmd.Verbose {
				fmt.Printf("\n[verbose] %s: %s\n", commandList, res)
			}
			for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
				if line != "" {
					if err := json.Unmarshal([]byte(line), &dockerServiceLs); err != nil {
						fmt.Println("🚫 Could not parse docker service ls JSON:\n", line)
						return err
					}
					if dockerServiceLs.Image == "" {
						return errors.New("service image not found")
					}
					commandPull.Args = append(commandPull.Args, ShellArg{
						Check: true,
						Value: shellescape.Quote(dockerServiceLs.Image),
					})
					return nil
				}
			}
			return errors.New("service not found")
		}).
		Run(commandPull.String(), func(res string) error {
			if cmd.Verbose {
				fmt.Printf("\n[verbose] %s: %s\n", commandPull, res)
			}
			return nil
		}).
		Run(commandUpdate, func(res string) error {
			if cmd.Verbose {
				fmt.Printf("\n[verbose] %s: %s\n", commandUpdate, res)
			}
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("🚫 Could not redeploy")
			}
			return err
		}).
		Error()
}

func (cmd *ServiceRedeployCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, cmd.Do)
	})
}
