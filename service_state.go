package rove

import "strings"

type ServiceState struct {
	Command             []string
	Env                 []string
	Image               string
	Init                bool
	Mounts              []string
	Networks            []string
	Publish             []string
	Replicas            string
	Secrets             []string
	UpdateDelay         string
	UpdateFailureAction string
	UpdateOrder         string
	UpdateParallelism   string
	User                string
	WorkDir             string
}

func (new *ServiceState) Diff(old *ServiceState) (string, DiffStatus) {
	lines := make([]DiffLine, 0)
	maxLeft := 0
	res := make([]string, len(lines))
	status := DiffSame

	lines, status = diffSlices(lines, status, "command", old.Command, new.Command)
	lines, status = diffSlices(lines, status, "env", old.Env, new.Env)
	lines, status = diffString(lines, status, "image", old.Image, new.Image)
	lines, status = diffBool(lines, status, "init", old.Init, new.Init)
	lines, status = diffSlices(lines, status, "mounts", old.Mounts, new.Mounts)
	lines, status = diffSlices(lines, status, "network", old.Networks, new.Networks)
	lines, status = diffSlices(lines, status, "publish", old.Publish, new.Publish)
	lines, status = diffString(lines, status, "replicas", old.Replicas, new.Replicas)
	lines, status = diffSlices(lines, status, "secret", old.Secrets, new.Secrets)
	lines, status = diffString(lines, status, "update-delay", old.UpdateDelay, new.UpdateDelay)
	lines, status = diffString(lines, status, "update-failure-action", old.UpdateFailureAction, new.UpdateFailureAction)
	lines, status = diffString(lines, status, "update-order", old.UpdateOrder, new.UpdateOrder)
	lines, status = diffString(lines, status, "update-parallelism", old.UpdateParallelism, new.UpdateParallelism)
	lines, status = diffString(lines, status, "user", old.User, new.User)
	lines, status = diffString(lines, status, "workdir", old.WorkDir, new.WorkDir)

	for _, line := range lines {
		if len(line.Left) > maxLeft {
			maxLeft = len(line.Left)
		}
	}
	for _, line := range lines {
		res = append(res, line.StringPadded(maxLeft))
	}
	return strings.Join(res, "\n"), status
}
