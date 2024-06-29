package rove

import (
	"slices"
	"testing"
)

func TestInspectCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{Result: `[{"ID":"fake-service-id","Version":{"Index":1234}}]`}
		expectedCmd := []string{"docker service inspect --format json fake-service"}
		expected := `{
    "ID": "fake-service-id",
    "Version": {
        "Index": 1234
    }
}` + "\n"

		capture(t).
			Run(func() error {
				cmd := &InspectCommand{
					Machine: "default",
					Name:    "fake-service",
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
