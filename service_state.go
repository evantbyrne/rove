package rove

import "strings"

type ServiceState struct {
	Command  []string
	Image    string
	Publish  []string
	Replicas string
}

func (new *ServiceState) Diff(old *ServiceState) (string, DiffStatus) {
	lines := make([]DiffLine, 0)
	maxLeft := 0
	res := make([]string, len(lines))
	status := DiffSame

	lines, status = diffSlices(lines, status, "command", old.Command, new.Command)
	lines, status = diffString(lines, status, "image", old.Image, new.Image)
	lines, status = diffSlices(lines, status, "ports", old.Publish, new.Publish)
	lines, status = diffString(lines, status, "replicas", old.Replicas, new.Replicas)

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