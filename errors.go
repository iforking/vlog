package vlog

type wrappedError struct {
	message string
	cause   error
}

func (we *wrappedError) Error() string {
	return we.message + ": " + we.cause.Error()
}

func wrapError(message string, err error) error {
	return &wrappedError{message: message, cause: err}
}
