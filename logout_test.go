package rove

import (
	"slices"
	"testing"
)

func TestLogoutCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		expectedCmd := []string{"docker logout example.com:8080"}
		expected := "\nRemote machine logged out of 'example.com:8080'.\n\n"

		capture(t).
			Run(func() error {
				cmd := &LogoutCommand{
					Machine:  "default",
					Registry: "example.com:8080",
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
