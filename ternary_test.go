package rove

import (
	"testing"
)

func TestTernary(t *testing.T) {
	resultStr := ternary(true, "truthy", "falsy")
	if resultStr != "truthy" {
		t.Errorf("'%#v' did not match expected.", resultStr)
	}

	resultInt := ternary(false, 1, 0)
	if resultInt != 0 {
		t.Errorf("'%#v' did not match expected.", resultInt)
	}
}
