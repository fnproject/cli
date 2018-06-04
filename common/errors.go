package common

// NotFoundError represents error string.
type NotFoundError struct {
	S string
}

func (e *NotFoundError) Error() string {
	return e.S
}

// NewNotFoundError returns a new error.
func NewNotFoundError(s string) *NotFoundError {
	return &NotFoundError{S: s}
}
