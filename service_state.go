package rove

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type DiffStatus string

const (
	DiffSame   DiffStatus = "DiffSame"
	DiffCreate DiffStatus = "DiffCreate"
	DiffDelete DiffStatus = "DiffDelete"
	DiffUpdate DiffStatus = "DiffUpdate"
)

type ServiceState struct {
	Command  []string
	Image    string
	Publish  []string
	Replicas string
}

func (new *ServiceState) Diff(old *ServiceState) (string, DiffStatus) {
	lines := make([]string, 0)
	status := DiffSame
	lines, status = diffSlices(lines, status, "command", old.Command, new.Command)
	lines, status = diffString(lines, status, "image", old.Image, new.Image)
	lines, status = diffSlices(lines, status, "publish", old.Publish, new.Publish)
	lines, status = diffString(lines, status, "replicas", old.Replicas, new.Replicas)
	return strings.Join(lines, "\n"), status
}

func m(value any) string {
	out, _ := json.Marshal(value)
	return string(out)
}

func diffSlices(lines []string, status DiffStatus, name string, old []string, new []string) ([]string, DiffStatus) {
	if slices.Equal(old, new) {
		if len(new) > 0 {
			lines = append(lines, fmt.Sprintf("     %s = %s", name, m(new)))
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, fmt.Sprintf(" +   %s = %s", name, m(new)))
		if status == DiffSame || status == DiffCreate {
			status = DiffCreate
		} else if status == DiffDelete {
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, fmt.Sprintf(" -   %s = %s", name, m(old)))
		if status == DiffSame || status == DiffDelete {
			status = DiffDelete
		} else if status == DiffCreate {
			status = DiffUpdate
		}
		return lines, DiffDelete
	}
	lines = append(lines, fmt.Sprintf(" -   %s = %s", name, m(old)))
	lines = append(lines, fmt.Sprintf(" +   %s = %s", name, m(new)))
	return lines, DiffUpdate
}

func diffString(lines []string, status DiffStatus, name string, old string, new string) ([]string, DiffStatus) {
	if old == new {
		if len(new) > 0 {
			lines = append(lines, fmt.Sprintf("     %s = %s", name, m(new)))
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, fmt.Sprintf(" +   %s = %s", name, m(new)))
		if status == DiffSame || status == DiffCreate {
			status = DiffCreate
		} else if status == DiffDelete {
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, fmt.Sprintf(" -   %s = %s", name, m(old)))
		if status == DiffSame || status == DiffDelete {
			status = DiffDelete
		} else if status == DiffCreate {
			status = DiffUpdate
		}
		return lines, DiffDelete
	}
	lines = append(lines, fmt.Sprintf(" -   %s = %s", name, m(old)))
	lines = append(lines, fmt.Sprintf(" +   %s = %s", name, m(new)))
	return lines, DiffUpdate
}
