package internal

// displayedError wraps an error that has already been reported to the user,
// so Cmd does not print it a second time.
type displayedError struct{ err error }

func (e *displayedError) Error() string { return e.err.Error() }

func (e *displayedError) Unwrap() error { return e.err }

// displayed marks err as already reported to the user.
func displayed(err error) error {
	if err == nil {
		return nil
	}

	return &displayedError{err: err}
}
