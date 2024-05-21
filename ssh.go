package rove

import (
	"bytes"
	"log"

	"golang.org/x/crypto/ssh"
)

func SshConnect(address string, user string, key []byte) *ssh.Client {
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
	return client
}

func SshRun(client *ssh.Client, command string) (string, error) {
	var b bytes.Buffer
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()
	session.Stdout = &b
	err = session.Run(command)
	return b.String(), err
}
