package jsonlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// custom level type to represent severity of log entry
type Level int8

const (
	LevelInfo  Level = 1
	LevelError Level = 2
	LevelFatal Level = 3
	LevelOff   Level = 4
)

// return a human-friendly string for severity level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelOff:
		return "OFF"
	default:
		return "UNKNOWN"
	}
}

// define a custom logger
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// return a new logger instance
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// helper methods for writing info messages to logger
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

// helper methods for writing error messages to logger
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

// helper methods for writing fatal messages to logger
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

// internal method for writing to logger
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// return without printing if severity is below minLevel
	if level < l.minLevel {
		return 0, nil
	}
	// declare anonymous struct to hold data for log entry
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// include stack trace for error and fatal entries
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// variable to hold actual log entry text
	var line []byte

	// marshal the aux struct
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(fmt.Sprintf("%s: unable to marshal log message: %s", LevelError.String(), err.Error()))
	}

	// lock mutex so that multiple log entries don't write concurrently
	l.mu.Lock()
	defer l.mu.Unlock()

	// write log entry followed by newline
	return l.out.Write(append(line, '\n'))
}

// implement Write method on our logger so that it satisfies the io.Writer interface
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
