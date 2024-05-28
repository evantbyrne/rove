package rove

import (
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
)

type ShellArg struct {
	Check bool
	Value string
}

func (arg ShellArg) String() string {
	if !arg.Check || arg.Value == "" {
		return ""
	}
	return arg.Value
}

type ShellCommand struct {
	Name  string
	Flags []ShellFlag
	Args  []ShellArg
}

func (cmd ShellCommand) String() string {
	parts := []string{cmd.Name}
	for _, flag := range cmd.Flags {
		if str := flag.String(); str != "" {
			parts = append(parts, str)
		}
	}
	for _, arg := range cmd.Args {
		if str := arg.String(); str != "" {
			parts = append(parts, str)
		}
	}
	return strings.Join(parts, " ")
}

type ShellFlag struct {
	Check bool
	Name  string
	Value string
}

func (flag ShellFlag) String() string {
	if !flag.Check || flag.Name == "" {
		return ""
	}
	if flag.Value == "" {
		return fmt.Sprint("--", flag.Name)
	}
	return fmt.Sprintf("--%s %s", flag.Name, shellescape.Quote(flag.Value))
}
