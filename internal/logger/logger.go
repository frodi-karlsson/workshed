package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	HELP
	SUCCESS
	ERROR
)

type LogFormat int

const (
	HUMAN LogFormat = iota
	JSON
	RAW
)

type Logger struct {
	level     LogLevel
	format    LogFormat
	command   string
	log       *log.Logger
	timestamp bool
	writer    io.Writer
}

type LoggerOption func(*Logger)

func WithTestOutput(writer io.Writer) LoggerOption {
	return func(l *Logger) {
		l.writer = writer
		l.log = log.New(writer, "", 0)
	}
}

func NewLogger(level LogLevel, command string, opts ...LoggerOption) *Logger {
	format := HUMAN
	if envFormat := os.Getenv("WORKSHED_LOG_FORMAT"); envFormat != "" {
		switch envFormat {
		case "json":
			format = JSON
		case "raw":
			format = RAW
		}
	}

	l := &Logger{
		level:     level,
		format:    format,
		command:   command,
		log:       log.New(os.Stdout, "", 0),
		timestamp: format == JSON,
		writer:    os.Stdout,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ERROR {
		l.logMessage("ERROR", msg, args)
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= INFO {
		l.logMessage("INFO", msg, args)
	}
}

func (l *Logger) Help(msg string, args ...interface{}) {
	if l.level <= HELP {
		l.logMessage("HELP", msg, args)
	}
}

func (l *Logger) Success(msg string, args ...interface{}) {
	if l.level <= SUCCESS {
		l.logMessage("SUCCESS", msg, args)
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logMessage("DEBUG", msg, args)
	}
}

// Internal log dispatcher
func (l *Logger) logMessage(level, msg string, args []interface{}) {
	switch l.format {
	case JSON:
		l.logJSON(level, msg, args)
	case RAW:
		l.logRaw(level, msg, args)
	default:
		l.logHuman(level, msg, args)
	}
}

// Human-readable format
func (l *Logger) logHuman(level, msg string, args []interface{}) {
	if l.command != "" {
		l.log.Printf("%s: %s | command: %s", level, msg, l.command)
	} else {
		l.log.Printf("%s: %s", level, msg)
	}

	if len(args) > 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				if _, ok := args[i+1].(string); ok {
					l.log.Printf("  %s: %q", args[i], args[i+1])
				} else {
					l.log.Printf("  %s: %v", args[i], args[i+1])
				}
			}
		}
	}
}

// JSON format for machine readability
func (l *Logger) logJSON(level, msg string, args []interface{}) {
	entry := map[string]interface{}{
		"level":   level,
		"message": msg,
		"command": l.command,
	}

	if l.timestamp {
		entry["timestamp"] = time.Now().Format(time.RFC3339)
	}

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			entry[fmt.Sprintf("%v", args[i])] = args[i+1]
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		l.log.Printf("JSON marshal error: %v", err)
		return
	}

	l.log.Printf("%s", string(data))
}

// Raw format for scripting
func (l *Logger) logRaw(level, msg string, args []interface{}) {
	l.log.Printf("%s", msg)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			l.log.Printf("%v", args[i+1])
		}
	}
}

// SafeFprintf safely writes to a writer, handling errors
func SafeFprintf(w interface {
	Write(p []byte) (n int, err error)
}, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// SafeFprintln safely writes to a writer, handling errors
func SafeFprintln(w interface {
	Write(p []byte) (n int, err error)
}, args ...interface{}) {
	_, _ = fmt.Fprintln(w, args...)
}
