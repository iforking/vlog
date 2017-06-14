package verylog

import (
	"os"
	"time"
	"path/filepath"
	"errors"
)

// Appender that write log to local file
type FileAppender struct {
	basePath    string
	currentFile *os.File // current opened file
	rotater     Rotater
	len         uint64 // data byte has been written
	recordNum   uint64 // records/lines has been written
	normal      bool
}

func NewFileAppender(path string) (Appender, error) {
	file, err := openFile(path)
	if err != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fileInfo.Size()
	}
	return &FileAppender{basePath: path, currentFile: file}, nil
}

func (f *FileAppender) Write(data string) (written int, err error) {
	f.len += uint64(len(data))
	f.recordNum += 1
	now := time.Now()
	if f.rotater != nil {
		shouldRotate, suffix := f.rotater.Check(f.len, f.recordNum, now)
		if shouldRotate {
			//rotate
			err := f.rotateFile(f.basePath + "-" + suffix)
			if err != nil {
				// rotate failed, stopping writing
				return 0, errors.New("rotate file error:" + err.Error())
			}
			f.len = 0
			f.recordNum = 0
		}
	}

	return f.currentFile.WriteString(data)
}

func (f *FileAppender) rotateFile(renamePath string) error {
	// should follow rename -> open new -> replace current -> close old steps.
	// would os.Rename work in windows when file is open? on windows should use FileShare.Delete when open file
	err := os.Rename(f.basePath, renamePath)
	if err != nil {
		return err
	}
	file, err := openFile(f.basePath)
	if err != nil {
		return err
	}
	oldFile := f.currentFile
	f.currentFile = file
	oldFile.Close()
	return nil
}

func openFile(path string) (*os.File, error) {
	if err := ensureParentPath(path); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
}

func ensureParentPath(path string) error {
	dir, _ := filepath.Split(path)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return nil
}

type Rotater interface {
	Check(bytes uint64, records uint64, timestamp time.Time) (shouldRotate bool, suffixName string)
}
