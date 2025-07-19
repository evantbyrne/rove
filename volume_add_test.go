package rove

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestVolumeAddCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		expectedCmd := []string{"docker volume create --group bar --label rove --name foo"}
		expected := fmt.Sprint(
			"\nRove will create the 'foo' volume.\n\n",
			"Confirmations skipped.\n\n",
			"Created 'foo' volume.\n\n")

		capture(t).
			Run(func() error {
				cmd := &VolumeAddCommand{
					Force:   true,
					Group:   "bar",
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
			"\nRove will create the 'foo' volume.\n\n",
			"Do you want Rove to run this deployment?\n",
			"  Type 'yes' to approve, or anything else to deny.\n",
			"  Enter a value: \n",
			"Created 'foo' volume.\n\n")

		capture(t).
			Run(func() error {
				cmd := &VolumeAddCommand{
					Group:   "bar",
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
