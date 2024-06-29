package rove

import (
	"os"
	"slices"
	"testing"
)

func TestLoginCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{}
		expectedCmd := []string{
			"cat fake-username.txt | docker login --username fake-username --password-stdin",
			"rm fake-username.txt",
		}
		expected := "\nRemote machine logged into 'docker.io' with username 'fake-username'.\n\n"

		capture(t).
			Run(func() error {
				f, err := os.CreateTemp("", "foo.txt")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(f.Name())
				defer f.Close()
				if _, err := f.Write([]byte("fake-secret-content")); err != nil {
					t.Fatal(err)
				}
				cmd := &LoginCommand{
					PasswordFile: f,
					Machine:      "default",
					Username:     "fake-username",
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
