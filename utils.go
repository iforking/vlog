package vlog

import (
	"path/filepath"
	"os"
)

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

func concatBytes(items ...[]byte) []byte {
	var l = 0
	for _, item := range items {
		l += len(item)
	}

	var result = make([]byte, l)
	var copied = 0
	for _, item := range items {
		copy(result[copied:], item)
		copied += len(item)
	}
	return result
}

// for files

func makeParentDirs(path string) error {
	dir, _ := filepath.Split(path)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return nil
}

//open file, create file and parent dirs if need
func openFile(path string) (*os.File, error) {
	if err := makeParentDirs(path); err != nil {
		return nil, wrapError("make file log parent dir failed", err)
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}
