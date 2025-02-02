package rove

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/evantbyrne/trance"
	"github.com/kballard/go-shellquote"
	"golang.org/x/crypto/ssh"
)

type LocalRunner struct {
	Err error
}

func (conn *LocalRunner) Error() error {
	return conn.Err
}

func (conn *LocalRunner) OnError(callback func(error) error) SshRunner {
	if conn.Err != nil {
		conn.Err = callback(conn.Err)
	}
	return conn
}

func (conn *LocalRunner) Run(command string, callback func(string) error) SshRunner {
	if conn.Err != nil {
		return conn
	}
	var bufferStdout bytes.Buffer
	words, err := shellquote.Split(command)
	if err != nil {
		conn.Err = fmt.Errorf("failed to shell split command '%s': %v", command, err)
		return conn
	}
	cmd := exec.Command(words[0], words[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &bufferStdout
	if err := cmd.Run(); err != nil {
		conn.Err = fmt.Errorf("failed to run command '%s': %v", command, err)
		return conn
	}
	conn.Err = callback(bufferStdout.String())
	return conn
}

func SshConnect(address string, user string, key []byte, callback func(conn SshRunner, stdin io.Reader) error) error {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer client.Close()
	return callback(&SshConnection{Client: client}, os.Stdin)
}

func SshMachine(machine *Machine, callback func(conn SshRunner, stdin io.Reader) error) error {
	key, err := os.ReadFile(machine.KeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key file: %v", err)
	}
	return SshConnect(fmt.Sprintf("%s:%d", machine.Address, machine.Port), machine.User, key, callback)
}

func SshMachineByName(local bool, name string, callback func(conn SshRunner, stdin io.Reader) error) error {
	if local {
		return callback(&LocalRunner{}, os.Stdin)
	}
	name = cmp.Or(name, GetPreference(DefaultMachine))
	if name == "" {
		return errors.New("🚫 No machine specified. Either run `rove machine use [NAME]` to set the default, or use the `--machine [NAME]` flag on individual commands")
	}
	return trance.Query[Machine]().
		Filter("name", "=", name).
		First().
		OnError(func(err error) error {
			if err != nil && errors.Is(err, trance.ErrorNotFound{}) {
				return fmt.Errorf("🚫 No machine with name '%s' configured", name)
			}
			return err
		}).
		Then(func(machine *Machine) error {
			return SshMachine(machine, callback)
		}).
		Error
}

type SshConnection struct {
	Client *ssh.Client
	Err    error
}

func (conn *SshConnection) Error() error {
	return conn.Err
}

func (conn *SshConnection) OnError(callback func(error) error) SshRunner {
	if conn.Err != nil {
		conn.Err = callback(conn.Err)
	}
	return conn
}

func (conn *SshConnection) Run(command string, callback func(string) error) SshRunner {
	if conn.Err != nil {
		return conn
	}
	var bufferStdout bytes.Buffer
	session, err := conn.Client.NewSession()
	if err != nil {
		conn.Err = fmt.Errorf("failed to create session for command '%s': %v", command, err)
		return conn
	}
	defer session.Close()
	session.Stderr = os.Stderr
	session.Stdout = &bufferStdout
	if err := session.Run(command); err != nil {
		conn.Err = fmt.Errorf("failed to run command '%s': %v", command, err)
		return conn
	}
	conn.Err = callback(bufferStdout.String())
	return conn
}

func confirmDeployment(force bool, stdin io.Reader) error {
	if force {
		fmt.Println("\nConfirmations skipped.")
	} else {
		fmt.Println("\nDo you want Rove to run this deployment?")
		fmt.Println("  Type 'yes' to approve, or anything else to deny.")
		fmt.Print("  Enter a value: ")
		line, err := bufio.NewReader(stdin).ReadString('\n')
		if err != nil {
			fmt.Println("🚫 Could not read from STDIN")
			return err
		}
		if strings.ToLower(strings.TrimSpace(line)) != "yes" {
			return errors.New("🚫 Deployment canceled because response did not match 'yes'")
		}
	}
	return nil
}

type SshRunner interface {
	Error() error
	OnError(func(error) error) SshRunner
	Run(string, func(string) error) SshRunner
}

type SshConnectionMock struct {
	CommandsRun []string
	Err         error
	Result      string
}

func (conn *SshConnectionMock) Error() error {
	return conn.Err
}

func (conn *SshConnectionMock) OnError(callback func(error) error) SshRunner {
	if conn.Err != nil {
		conn.Err = callback(conn.Err)
	}
	return conn
}

func (conn *SshConnectionMock) Run(command string, callback func(string) error) SshRunner {
	if conn.Err != nil {
		return conn
	}
	conn.CommandsRun = append(conn.CommandsRun, command)
	conn.Err = callback(conn.Result)
	return conn
}
