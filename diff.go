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

func diffSlices(lines []DiffLine, status DiffStatus, name string, old []string, new []string) ([]DiffLine, DiffStatus) {
	if slices.Equal(old, new) {
		if len(new) > 0 {
			lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffSame})
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffCreate})
		if status == DiffSame || status == DiffCreate {
			status = DiffCreate
		} else if status == DiffDelete {
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(old), Status: DiffDelete})
		if status == DiffSame || status == DiffDelete {
			status = DiffDelete
		} else if status == DiffCreate {
			status = DiffUpdate
		}
		return lines, DiffDelete
	}
	lines = append(lines, DiffLine{Left: name, Right: mustMarshal(old), Status: DiffDelete})
	lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffCreate})
	return lines, DiffUpdate
}

func diffString(lines []DiffLine, status DiffStatus, name string, old string, new string) ([]DiffLine, DiffStatus) {
	if old == new {
		if len(new) > 0 {
			lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffSame})
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffCreate})
		if status == DiffSame || status == DiffCreate {
			status = DiffCreate
		} else if status == DiffDelete {
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(old), Status: DiffDelete})
		if status == DiffSame || status == DiffDelete {
			status = DiffDelete
		} else if status == DiffCreate {
			status = DiffUpdate
		}
		return lines, DiffDelete
	}
	lines = append(lines, DiffLine{Left: name, Right: mustMarshal(old), Status: DiffDelete})
	lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffCreate})
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

func mustMarshal(value any) string {
	out, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(out)
}
