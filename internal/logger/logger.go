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
	WARN
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

func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= WARN {
		l.logMessage("WARN", msg, args)
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
		hintFound := false
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				key := args[i]
				value := args[i+1]

				if key == "hint" {
					if !hintFound {
						l.log.Printf("  hint: %v", value)
						hintFound = true
					}
				} else if _, ok := value.(string); ok {
					l.log.Printf("  %s: %q", key, value)
				} else {
					l.log.Printf("  %s: %v", key, value)
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

// UncheckedFprintf writes to a writer without checking for errors.
//
// This function intentionally discards write errors. It is appropriate for
// non-critical output such as help text, usage messages, and hints where
// write failures are considered terminal conditions (e.g., broken pipe).
// Use fmt.Fprintf directly if you need to handle write errors.
func UncheckedFprintf(w interface {
	Write(p []byte) (n int, err error)
}, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w, format, args...)
}
