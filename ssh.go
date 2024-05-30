package rove

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/evantbyrne/trance"
	"golang.org/x/crypto/ssh"
)

func SshConnect(address string, user string, key []byte, callback func(conn *SshConnection) error) error {
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
	return callback(&SshConnection{Client: client})
}

func SshMachine(machine *Machine, callback func(conn *SshConnection) error) error {
	key, err := os.ReadFile(machine.KeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key file: %v", err)
	}
	return SshConnect(fmt.Sprintf("%s:%d", machine.Address, machine.Port), machine.User, key, callback)
}

func SshMachineByName(name string, callback func(conn *SshConnection) error) error {
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
	Error  error
}

func (conn *SshConnection) OnError(callback func(error) error) *SshConnection {
	if conn.Error != nil {
		conn.Error = callback(conn.Error)
	}
	return conn
}

func (conn *SshConnection) Run(command string, callback func(string) error) *SshConnection {
	if conn.Error != nil {
		return conn
	}
	var bufferStdout bytes.Buffer
	session, err := conn.Client.NewSession()
	if err != nil {
		conn.Error = fmt.Errorf("failed to create session for command '%s': %v", command, err)
		return conn
	}
	defer session.Close()
	session.Stderr = os.Stderr
	session.Stdout = &bufferStdout
	if err := session.Run(command); err != nil {
		conn.Error = fmt.Errorf("failed to run command '%s': %v", command, err)
		return conn
	}
	conn.Error = callback(bufferStdout.String())
	return conn
}
