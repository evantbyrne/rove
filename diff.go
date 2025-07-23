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

func diffBool(lines []DiffLine, status DiffStatus, name string, old bool, new bool) ([]DiffLine, DiffStatus) {
	if new {
		if !old {
			lines = append(lines, DiffLine{Left: name, Right: "true", Status: DiffCreate})
			switch status {
			case DiffSame:
				status = DiffCreate
			case DiffDelete:
				status = DiffUpdate
			}
		} else {
			lines = append(lines, DiffLine{Left: name, Right: "true", Status: DiffSame})
		}
	} else {
		if old {
			lines = append(lines, DiffLine{Left: name, Right: "true", Status: DiffDelete})
			switch status {
			case DiffSame:
				status = DiffDelete
			case DiffCreate:
				status = DiffUpdate
			}
		}
	}
	return lines, status
}

func diffSlices(lines []DiffLine, status DiffStatus, name string, old []string, new []string) ([]DiffLine, DiffStatus) {
	if slices.Equal(old, new) {
		if len(new) > 0 {
			lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffSame})
		}
		return lines, status
	}
	if len(old) == 0 && len(new) > 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(new), Status: DiffCreate})
		switch status {
		case DiffSame, DiffCreate:
			status = DiffCreate
		case DiffDelete:
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(old), Status: DiffDelete})
		switch status {
		case DiffSame, DiffDelete:
			status = DiffDelete
		case DiffCreate:
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
		switch status {
		case DiffSame, DiffCreate:
			status = DiffCreate
		case DiffDelete:
			status = DiffUpdate
		}
		return lines, status
	}
	if len(old) > 0 && len(new) == 0 {
		lines = append(lines, DiffLine{Left: name, Right: mustMarshal(old), Status: DiffDelete})
		switch status {
		case DiffSame, DiffDelete:
			status = DiffDelete
		case DiffCreate:
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
