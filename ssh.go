package rove

import (
	"bytes"
	"fmt"
	"log"

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

type SshConnection struct {
	Client *ssh.Client
	Error  error
}

func (conn *SshConnection) Run(command string, callback func(string) error) *SshConnection {
	if conn.Error != nil {
		return conn
	}
	var b bytes.Buffer
	session, err := conn.Client.NewSession()
	if err != nil {
		conn.Error = fmt.Errorf("failed to create session for command '%s': %v", command, err)
		return conn
	}
	defer session.Close()
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		conn.Error = fmt.Errorf("failed to run command '%s': %v", command, err)
		return conn
	}
	conn.Error = callback(b.String())
	return conn
}
