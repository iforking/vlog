package vlog

import (
	"os"
	"time"
	"path/filepath"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"unsafe"
	"fmt"
	"strconv"
)

// Appender that write log to local file
type FileAppender struct {
	path    string
	file    unsafe.Pointer //*os.File, current opened file
	rotater Rotater
	normal  bool
}

// create new file appender.
// path is the base path and filename of log file.
// appender can be nil, then the file would not be rotated.
func NewFileAppender(path string, rotater Rotater) (Appender, error) {
	file, err := openFile(path)
	if err != nil {
		return nil, err
	}

	if rotater != nil {
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		suffixes := getLogSuffixed(path)
		rotater.setInitStatus(fileInfo.ModTime(), fileInfo.Size(), suffixes)
	}
	return &FileAppender{path: path, file: unsafe.Pointer(file), rotater: rotater}, nil
}

func (f *FileAppender) Write(data string) (written int, err error) {
	if f.rotater != nil {
		shouldRotate, suffix := f.rotater.Check(time.Now(), len(data), 1)
		if shouldRotate {
			//rotate
			ext := filepath.Ext(f.path)
			base := f.path[:len(f.path)-len(ext)]
			err := f.rotateFile(base + "." + suffix + ext)
			if err != nil {
				// rotate failed, still use the current file?
				print("rotate failed, stopping writing")
			}
		}
	}

	return f.currentFile().WriteString(data)
}

func (f *FileAppender) currentFile() *os.File {
	return (*os.File)(atomic.LoadPointer(&f.file))
}

func (f *FileAppender) swapFile(oldFile *os.File, file *os.File) bool {
	return atomic.CompareAndSwapPointer(&f.file, unsafe.Pointer(oldFile), unsafe.Pointer(file))
}

func (f *FileAppender) rotateFile(renamePath string) error {
	// should follow rename -> open new -> replace current -> close old steps.
	// would os.Rename work in windows when file is open? on windows should use FileShare.Delete when open file
	err := os.Rename(f.path, renamePath)
	if err != nil {
		return err
	}
	file, err := openFile(f.path)
	if err != nil {
		return err
	}
	oldFile := f.currentFile()
	if f.swapFile(oldFile, file) {
		oldFile.Close()
	} else {
		//should not happen if appender act rightly ?
	}
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

func getLogSuffixed(path string) []string {
	dir, filename := filepath.Split(path)
	extension := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(extension)]
	files, _ := ioutil.ReadDir(dir)

	var suffixes []string
	for _, f := range files { // ignore error
		logFileName := f.Name()
		if !strings.HasPrefix(logFileName, baseName) {
			continue
		}
		remain := logFileName[len(baseName):]
		if len(remain) == 0 {
			continue
		}
		idx := strings.Index(remain, extension)
		if idx <= 1 {
			continue
		}
		suffix := remain[1:idx]
		suffixes = append(suffixes, suffix)
	}
	return suffixes
}

type Rotater interface {
	// tell rotater init log file status, so rotater can determine when and how to do next rotate.
	// param lastModify is the modify time of last logfile
	// param size is the initial log file size; if create new file, this param is 0.
	// param suffixes is the existed suffixes of log files in current log directory.
	setInitStatus(lastModify time.Time, size int64, suffixes []string)

	// call this to determine if should do rotate.
	// timestamp is the time the last log logged;
	// bytes is the data size logged since last call to this method;
	// records is new log num since last call to this method;
	Check(timestamp time.Time, bytes int, records int) (shouldRotate bool, suffixName string)
}

// Rotate log file by time
type TimeRotater struct {
	duration     time.Duration
	suffixFormat string
	last         unsafe.Pointer // *time.Time
}

// create rotater rotate log by time
func NewTimeRotater(duration time.Duration, suffixFormat string) Rotater {
	return &TimeRotater{duration: duration, suffixFormat: suffixFormat}
}

// create rotater rotate log every hour
func NewHourRotater() Rotater {
	return NewTimeRotater(time.Hour, "2006-01-02-15")
}

// create rotater rotate log every day
func NewDayRotater() Rotater {
	return NewTimeRotater(time.Hour*24, "2006-01-02")
}

func (t *TimeRotater) lastTime() *time.Time {
	return (*time.Time)(atomic.LoadPointer(&t.last))
}

func (t *TimeRotater) setLastTime(ts *time.Time) {
	atomic.StorePointer(&t.last, unsafe.Pointer(ts))
}

func (t *TimeRotater) casLastTime(oldTime *time.Time, ts *time.Time) bool {
	return atomic.CompareAndSwapPointer(&t.last, unsafe.Pointer(oldTime), unsafe.Pointer(ts))
}

func (t *TimeRotater) Check(timestamp time.Time, bytes int, records int) (shouldRotate bool, suffixName string) {
	intervalSeconds := int64(t.duration / time.Second)
	last := t.lastTime()
	diff := timestamp.Unix()/intervalSeconds - last.Unix()/intervalSeconds
	if diff <= 0 {
		return false, ""
	}
	suffix := last.Format(t.suffixFormat)
	if t.casLastTime(last, &timestamp) {
		return true, suffix
	}
	return false, ""
}

func (t *TimeRotater) setInitStatus(lastModify time.Time, size int64, suffixes []string) {
	t.setLastTime(&lastModify)
}

// rotate based on file size
type SizeRotater struct {
	rotateSize  int64
	size        int64
	seq         int64
	SuffixWidth int
}

// create file size rotater, rotate log file when file size larger than rotateSize, in bytes
func NewSizeRotater(rotateSize int64) Rotater {
	return &SizeRotater{rotateSize: rotateSize, SuffixWidth: 5}
}

func (sr *SizeRotater) setInitStatus(lastModify time.Time, size int64, suffixes []string) {
	sr.setSize(size)
	var maxSeq = 0
	for _, suffix := range suffixes {
		if seq, err := strconv.Atoi(suffix); err == nil {
			if seq > maxSeq {
				maxSeq = seq
			}
		}
	}
	sr.setSeq(int64(maxSeq))
}

func (sr *SizeRotater) Check(timestamp time.Time, bytes int, records int) (shouldRotate bool, suffixName string) {
	size := sr.addSize(int64(bytes))
	if size < sr.rotateSize {
		return false, ""
	}
	if sr.casSize(size, 0) {
		seq := sr.increaseSeq()
		seqStr := fmt.Sprintf("%0"+strconv.Itoa(sr.SuffixWidth)+"d", seq)
		return true, seqStr
	}
	return false, ""
}

func (sr *SizeRotater) loadSize() int64 {
	return atomic.LoadInt64(&sr.size)
}
func (sr *SizeRotater) setSize(size int64) {
	atomic.StoreInt64(&sr.size, size)
}
func (sr *SizeRotater) casSize(oldSize int64, size int64) bool {
	return atomic.CompareAndSwapInt64(&sr.size, oldSize, size)
}
func (sr *SizeRotater) addSize(delta int64) int64 {
	return atomic.AddInt64(&sr.size, delta)
}
func (sr *SizeRotater) setSeq(seq int64) {
	atomic.StoreInt64(&sr.seq, seq)
}
func (sr *SizeRotater) increaseSeq() int64 {
	return atomic.AddInt64(&sr.seq, 1)
}
