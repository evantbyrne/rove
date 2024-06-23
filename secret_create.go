package rove

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/pkg/sftp"
)

type SecretCreateCommand struct {
	Name string   `arg:"" name:"name" help:"Name of secret."`
	File *os.File `arg:"" name:"file" help:"Secret file. Use dash (-) to read from STDIN."`

	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

type SecretCreateJson struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (cmd *SecretCreateCommand) Do(conn SshRunner, stdin io.Reader) error {
	fileName := cmd.Name + ".txt"
	secret, err := io.ReadAll(cmd.File)
	if err != nil {
		return err
	}

	if connReal, ok := conn.(*SshConnection); ok {
		transfer, err := sftp.NewClient(connReal.Client)
		if err != nil {
			return err
		}
		defer transfer.Close()

		fh, err := transfer.Create(fileName)
		if err != nil {
			return err
		}
		if _, err := fh.Write([]byte(secret)); err != nil {
			return err
		}
		fh.Close()
	}

	err = conn.
		Run(fmt.Sprintf("docker secret create --label 'rove=secret' %s %s", shellescape.Quote(cmd.Name), fileName), func(res string) error {
			id := strings.TrimSpace(res)
			if cmd.Json {
				output := SecretCreateJson{
					Id:   id,
					Name: cmd.Name,
				}
				out, err := json.MarshalIndent(output, "", "    ")
				if err != nil {
					fmt.Println("🚫 Could not format JSON:\n", output)
					return err
				}
				fmt.Println(string(out))
			} else {
				fmt.Printf("\nRove created the '%s' secret with ID '%s'.\n\n", cmd.Name, id)
			}
			return nil
		}).
		OnError(func(err error) error {
			if err != nil {
				fmt.Println("\n🚫 Could not add secret")
			}
			return err
		}).
		Error()

	conn.Run(fmt.Sprintf("rm %s", fileName), func(_ string) error {
		return nil
	})

	return err
}

func (cmd *SecretCreateCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, cmd.Do)
	})
}
