package rove

type ErrorSkip struct {
	Message string
}

func (err ErrorSkip) Error() string {
	if err.Message != "" {
		return err.Message
	}
	return "Skipped"
}
