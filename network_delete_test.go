package rove

import (
	"fmt"
	"slices"
	"testing"
)

func TestNetworkDeleteCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		expectedCmd := []string{"docker network rm foo"}
		expected := fmt.Sprint(
			"\nRove will delete the 'foo' network.\n\n",
			"Confirmations skipped.\n\n",
			"Deleted 'foo' network.\n\n")

		capture(t).
			Run(func() error {
				cmd := &NetworkDeleteCommand{
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
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
