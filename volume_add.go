package rove

import (
	"fmt"
	"io"
)

type VolumeAddCommand struct {
	Name string `arg:"" name:"name" help:"Name of volume."`

	Availability  string   `flag:"" name:"availability" help:"Cluster Volume availability (active, pause, drain)."`
	Driver        string   `flag:"" name:"driver" help:"Specify volume driver name."`
	Group         string   `flag:"" name:"group" help:"Cluster Volume group (cluster volumes)."`
	LimitBytes    string   `flag:"" name:"limit-bytes" help:"Minimum size of the Cluster Volume in bytes."`
	Opt           []string `flag:"" name:"opt" sep:"none" help:"Set driver specific options."`
	RequiredBytes string   `flag:"" name:"required-bytes" help:"Maximum size of the Cluster Volume in bytes."`
	Sharing       string   `flag:"" name:"sharing" help:"Cluster Volume access sharing (none, readonly, onewriter, all)."`
	Scope         string   `flag:"" name:"scope" help:"Cluster Volume access scope (single, multi)."`
	Type          string   `flag:"" name:"type" help:"Cluster Volume access type (mount, block)."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Force      bool   `flag:"" name:"force" help:"Skip confirmations."`
	Local      bool   `flag:"" name:"local" help:"Skip SSH and run on local machine."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

func (cmd *VolumeAddCommand) Do(conn SshRunner, stdin io.Reader) error {
	fmt.Printf("\nRove will create the '%s' volume.\n", cmd.Name)
	if err := confirmDeployment(cmd.Force, stdin); err != nil {
		return err
	}
	command := ShellCommand{
		Name: "docker volume create",
		Flags: []ShellFlag{
			{
				Check: cmd.Driver != "",
				Name:  "driver",
				Value: cmd.Driver,
			},
			{
				Check: cmd.Group != "",
				Name:  "group",
				Value: cmd.Group,
			},
			{
				Check: true,
				Name:  "label",
				Value: "rove",
			},
			{
				Check: cmd.LimitBytes != "",
				Name:  "limit-bytes",
				Value: cmd.LimitBytes,
			},
			{
				Check: true,
				Name:  "name",
				Value: cmd.Name,
			},
			{
				Check: cmd.RequiredBytes != "",
				Name:  "requiredBytes-bytes",
				Value: cmd.RequiredBytes,
			},
			{
				Check: cmd.Sharing != "",
				Name:  "sharing",
				Value: cmd.Sharing,
			},
			{
				Check: cmd.Scope != "",
				Name:  "scope",
				Value: cmd.Scope,
			},
			{
				Check: cmd.Type != "",
				Name:  "type",
				Value: cmd.Type,
			},
		},
	}
	for _, opt := range cmd.Opt {
		command.Flags = append(command.Flags, ShellFlag{
			Check: true,
			Name:  "opt",
			Value: opt,
		})
	}
	return conn.
		Run(command.String(), func(res string) error {
			fmt.Printf("\nCreated '%s' volume.\n\n", cmd.Name)
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("🚫 Could not add volume")
			}
			return err
		}).
		Error()
}

func (cmd *VolumeAddCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Local, cmd.Machine, cmd.Do)
	})
}
