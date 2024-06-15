package rove

import "strings"

type ServiceState struct {
	Command  []string
	Image    string
	Mounts   []string
	Networks []string
	Publish  []string
	Replicas string
	Secrets  []string
}

func (new *ServiceState) Diff(old *ServiceState) (string, DiffStatus) {
	lines := make([]DiffLine, 0)
	maxLeft := 0
	res := make([]string, len(lines))
	status := DiffSame

	lines, status = diffSlices(lines, status, "command", old.Command, new.Command)
	lines, status = diffString(lines, status, "image", old.Image, new.Image)
	lines, status = diffSlices(lines, status, "mounts", old.Mounts, new.Mounts)
	lines, status = diffSlices(lines, status, "networks", old.Networks, new.Networks)
	lines, status = diffSlices(lines, status, "ports", old.Publish, new.Publish)
	lines, status = diffString(lines, status, "replicas", old.Replicas, new.Replicas)
	lines, status = diffSlices(lines, status, "secrets", old.Secrets, new.Secrets)

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
