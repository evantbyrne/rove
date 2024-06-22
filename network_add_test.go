package rove

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func captureStdout(callback func() error) (string, error) {
	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	if err := callback(); err != nil {
		return "", err
	}
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	os.Stdout = rescueStdout
	return string(out), nil
}

func TestNetworkAddCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		stdout, err := captureStdout(func() error {
			cmd := &NetworkAddCommand{
				Force:   true,
				Machine: "default",
				Name:    "foo",
			}
			return cmd.Do(mock)
		})
		if err != nil {
			return err
		}
		expected := "docker network create --attachable --driver overlay --label rove --scope swarm foo"
		if len(mock.CommandsRun) != 1 || mock.CommandsRun[0] != expected {
			t.Fatalf("'%#v' did not match expected.", mock.CommandsRun)
		}
		expected = fmt.Sprint(
			"\nRove will create the 'foo' network.\n\n",
			"Confirmations skipped.\n",
			"Created 'foo' network.\n\n")
		if stdout != expected {
			t.Fatalf("'%s' did not match:\n'%s'", stdout, expected)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
