package rove

import "errors"

type ErrorSkip struct {
	Message string
}

func (err ErrorSkip) Error() string {
	if err.Message != "" {
		return err.Message
	}
	return "Skipped"
}

func SkipReset(err error) error {
	if errors.Is(err, ErrorSkip{}) {
		return nil
	}
	return err
}
