package rove

import (
	"slices"
	"testing"
)

func TestInspectCommand(t *testing.T) {
	if err := testDatabase(func() error {
		mock := &SshConnectionMock{Result: `[{"ID":"fake-service-id","Spec":{"Name":"files","Labels":{"rove":"service"},"Mode":{"Replicated":{"Replicas":1}},"UpdateConfig":{"Parallelism":1,"FailureAction":"pause","Monitor":5000000000,"MaxFailureRatio":0,"Order":"stop-first"},"TaskTemplate":{"ContainerSpec":{"Image":"python:3.12@sha256:05855f5bf06f5a004b0c1a8aaac73a9d9ea54390fc289d3e80ef52c4f90d5585","Args":["python3","-m","http.server","80"],"Init":false,"StopGracePeriod":10000000000,"DNSConfig":{},"Isolation":"default"}}},"Version":{"Index":1234}}]`}
		expectedCmd := []string{"docker service inspect --format json fake-service"}
		expected := "\n" + `Current state of fake-service:

   service fake-service:
     command  = ["python3","-m","http.server","80"]
     image    = "python:3.12"
     replicas = "1"` + "\n\n"

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

func TestInspectCommandJson(t *testing.T) {
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
					Json:    true,
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
