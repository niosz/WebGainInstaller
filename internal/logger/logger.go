package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level string

const (
	INFO    Level = "INFO"
	WARNING Level = "WARNING"
	ERROR   Level = "ERROR"
)

var (
	file *os.File
	mu   sync.Mutex
)

func Init(rootPath string) error {
	mu.Lock()
	defer mu.Unlock()
	var err error
	file, err = os.Create(filepath.Join(rootPath, "log.txt"))
	if err != nil {
		return fmt.Errorf("impossibile creare log.txt: %w", err)
	}
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		file.Close()
		file = nil
	}
}

func write(level Level, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		return
	}
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(file, "[%s](%s) %s\n", ts, level, msg)
}

func Info(format string, args ...interface{}) {
	write(INFO, format, args...)
}

func Warn(format string, args ...interface{}) {
	write(WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	write(ERROR, format, args...)
}
