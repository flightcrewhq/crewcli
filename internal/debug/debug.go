package debug

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	logger *logWriter
)

type logWriter struct {
	w *os.File
}

func Enable(fn string) (func(), error) {
	if err := os.MkdirAll(filepath.Base(fn), 0755); err != nil {
		return nil, err
	}

	if _, err := os.Stat(fn); err == nil {
		if err := os.Remove(fn); err != nil {
			return nil, err
		}
	}

	f, err := os.Create(fn)
	if err != nil {
		return nil, err
	}

	f.Sync()

	logger = &logWriter{
		w: f,
	}

	return func() {
		f.Close()
	}, nil
}

func Output(detail string, args ...interface{}) {
	if logger == nil {
		return
	}

	str := fmt.Sprintf(detail+"\n", args...)
	logger.w.WriteString(str)
	logger.w.Sync()
}
