// pkg/logging/logging.go
package logging

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

func NewLogger(output io.Writer) *Logger {
	return &Logger{
		Info:  log.New(output, "[INFO] ", log.Ldate|log.Ltime),
		Warn:  log.New(output, "[WARN] ", log.Ldate|log.Ltime),
		Error: log.New(output, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

func NewDefaultLogger() *Logger {
	return NewLogger(os.Stdout)
}
