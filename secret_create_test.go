package rove

import (
	"os"
	"slices"
	"testing"
)

func TestSecretCreateCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{Result: "fake-id\n"}
		expectedCmd := []string{
			"docker secret create --label 'rove=secret' foo foo.txt",
			"rm foo.txt",
		}
		expected := "\nRove created the 'foo' secret with ID 'fake-id'.\n\n"

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
				cmd := &SecretCreateCommand{
					File:    f,
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
