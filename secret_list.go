package rove

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DockerSecretLsJson struct {
	CreatedAt string `json:"CreatedAt"`
	Id        string `json:"ID"`
	Name      string `json:"Name"`
	UpdatedAt string `json:"UpdatedAt"`
}

type SecretListCommand struct {
	ConfigFile string `flag:"" name:"config" help:"Config file." type:"path" default:".rove"`
	Json       bool   `flag:"" name:"json" help:"Output as JSON."`
	Machine    string `flag:"" name:"machine" help:"Name of machine." default:""`
}

type SecretListJson struct {
	Secrets []DockerSecretLsJson
}

func (cmd *SecretListCommand) Run() error {
	return Database(cmd.ConfigFile, func() error {
		return SshMachineByName(cmd.Machine, func(conn SshRunner) error {
			return conn.
				Run("docker secret ls --format json --filter label=rove", func(res string) error {
					output := make([]DockerSecretLsJson, 0)
					for _, line := range strings.Split(strings.ReplaceAll(res, "\r\n", "\n"), "\n") {
						if line != "" {
							var dockerSecretLs DockerSecretLsJson
							if err := json.Unmarshal([]byte(line), &dockerSecretLs); err != nil {
								fmt.Println("🚫 Could not parse docker secret ls JSON:\n", line)
								return err
							}
							output = append(output, dockerSecretLs)
						}
					}
					if cmd.Json {
						t := SecretListJson{
							Secrets: output,
						}
						out, err := json.MarshalIndent(t, "", "    ")
						if err != nil {
							fmt.Println("🚫 Could not format JSON:\n", t)
							return err
						}
						fmt.Println(string(out))
					} else {
						for _, dockerSecretLs := range output {
							fmt.Println(dockerSecretLs.Id, dockerSecretLs.Name, dockerSecretLs.CreatedAt, dockerSecretLs.UpdatedAt)
						}
					}
					return nil
				}).
				Error()
		})
	})
}
