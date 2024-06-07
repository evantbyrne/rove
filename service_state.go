package rove

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type DiffLine struct {
	Left   string
	Right  string
	Status DiffStatus
}

func (line DiffLine) StringPadded(maxLeft int) string {
	pad := strings.Repeat(" ", maxLeft-len(line.Left))
	symbol := diffSymbol(line.Status)
	return fmt.Sprintf(" %s   %s = %s", symbol, line.Left+pad, line.Right)
}

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

func m(value any) string {
	out, _ := json.Marshal(value)
	return string(out)
}

func diffSlices(lines []DiffLine, status DiffStatus, name string, old []string, new []string) ([]DiffLine, DiffStatus) {
	if slices.Equal(old, new) {
		if len(new) > 0 {
			lines = append(lines, DiffLine{Left: name, Right: m(new), Status: DiffSame})
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, DiffLine{Left: name, Right: m(new), Status: DiffCreate})
		if status == DiffSame || status == DiffCreate {
			status = DiffCreate
		} else if status == DiffDelete {
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, DiffLine{Left: name, Right: m(old), Status: DiffDelete})
		if status == DiffSame || status == DiffDelete {
			status = DiffDelete
		} else if status == DiffCreate {
			status = DiffUpdate
		}
		return lines, DiffDelete
	}
	lines = append(lines, DiffLine{Left: name, Right: m(old), Status: DiffDelete})
	lines = append(lines, DiffLine{Left: name, Right: m(new), Status: DiffCreate})
	return lines, DiffUpdate
}

func diffString(lines []DiffLine, status DiffStatus, name string, old string, new string) ([]DiffLine, DiffStatus) {
	if old == new {
		if len(new) > 0 {
			lines = append(lines, DiffLine{Left: name, Right: m(new), Status: DiffSame})
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, DiffLine{Left: name, Right: m(new), Status: DiffCreate})
		if status == DiffSame || status == DiffCreate {
			status = DiffCreate
		} else if status == DiffDelete {
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, DiffLine{Left: name, Right: m(old), Status: DiffDelete})
		if status == DiffSame || status == DiffDelete {
			status = DiffDelete
		} else if status == DiffCreate {
			status = DiffUpdate
		}
		return lines, DiffDelete
	}
	lines = append(lines, DiffLine{Left: name, Right: m(old), Status: DiffDelete})
	lines = append(lines, DiffLine{Left: name, Right: m(new), Status: DiffCreate})
	return lines, DiffUpdate
}

func diffSymbol(status DiffStatus) string {
	switch status {
	case DiffCreate:
		return "+"
	case DiffDelete:
		return "-"
	case DiffUpdate:
		return "~"
	}
	return " "
}
