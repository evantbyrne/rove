package rove

import (
	"fmt"
	"slices"
	"testing"
)

func TestVolumeDeleteCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		expectedCmd := []string{"docker volume rm foo"}
		expected := fmt.Sprint(
			"\nRove will delete the 'foo' volume.\n\n",
			"Confirmations skipped.\n\n",
			"Deleted 'foo' volume.\n\n")

		capture(t).
			Run(func() error {
				cmd := &VolumeDeleteCommand{
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
