package rove

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestNetworkAddCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		expectedCmd := []string{"docker network create --attachable --driver overlay --label rove --scope swarm foo"}
		expected := fmt.Sprint(
			"\nRove will create the 'foo' network.\n\n",
			"Confirmations skipped.\n\n",
			"Created 'foo' network.\n\n")

		capture(t).
			Run(func() error {
				cmd := &NetworkAddCommand{
					Force:   true,
					Machine: "default",
					Name:    "foo",
				}
				return cmd.Do(mock, nil)
			}).
			ExpectStdout(expected)

		if !slices.Equal(mock.CommandsRun, expectedCmd) {
			t.Errorf("'%#v' did not match expected.", mock.CommandsRun)
		}

		expected = fmt.Sprint(
			"\nRove will create the 'foo' network.\n\n",
			"Do you want Rove to run this deployment?\n",
			"  Type 'yes' to approve, or anything else to deny.\n",
			"  Enter a value: \n",
			"Created 'foo' network.\n\n")

		capture(t).
			Run(func() error {
				cmd := &NetworkAddCommand{
					Machine: "default",
					Name:    "foo",
				}
				return cmd.Do(mock, strings.NewReader("yes\n"))
			}).
			ExpectStdout(expected)

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
