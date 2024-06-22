package rove

import (
	"io"
	"os"
	"testing"
)

type captured struct {
	Stdout string
	t      *testing.T
}

func (c *captured) ExpectStdout(expected string) *captured {
	if c.Stdout != expected {
		c.t.Errorf("'%s' STDOUT did not match:\n'%s'", c.Stdout, expected)
	}
	return c
}

func (c *captured) Run(callback func() error) *captured {
	savedStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		c.t.Fatal(err)
	}
	os.Stdout = w
	if err := callback(); err != nil {
		c.t.Fatal(err)
	}
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		c.t.Fatal(err)
	}
	os.Stdout = savedStdout
	c.Stdout = string(out)
	return c
}

func capture(t *testing.T) *captured {
	return &captured{
		t: t,
	}
}
