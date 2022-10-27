package debug

import (
	"fmt"
	"os"
)

var (
	logger *logWriter
)

type logWriter struct {
	w *os.File
}

func init() {
	fn := "tmp/debug.log"
	if err := os.MkdirAll("tmp", 0755); err != nil {
		panic(err)
	}

	if _, err := os.Stat(fn); err == nil {
		if err := os.Remove(fn); err != nil {
			panic(err)
		}
	}

	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}

	f.WriteString("Hello\n")
	f.Sync()

	logger = &logWriter{
		w: f,
	}
}

func Output(detail string, args ...interface{}) {
	str := fmt.Sprintf(detail+"\n", args...)
	logger.w.WriteString(str)
	if err := logger.w.Sync(); err != nil {
		panic(err)
	}
}
