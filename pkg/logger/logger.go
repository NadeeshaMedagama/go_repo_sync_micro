package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

// Logger provides structured logging
type Logger struct {
	level      LogLevel
	fileWriter io.Writer
	prefix     string
}

var (
	defaultLogger *Logger
	levelStrings  = map[LogLevel]string{
		DEBUG:   "DEBUG",
		INFO:    "INFO",
		WARNING: "WARN",
		ERROR:   "ERROR",
	}
)

// Init initializes the default logger
func Init(level, logFilePath, service string) error {
	logLevel := parseLogLevel(level)

	// Ensure log directory exists
	if logFilePath != "" {
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		defaultLogger = &Logger{
			level:      logLevel,
			fileWriter: io.MultiWriter(os.Stdout, file),
			prefix:     service,
		}
	} else {
		defaultLogger = &Logger{
			level:      logLevel,
			fileWriter: os.Stdout,
			prefix:     service,
		}
	}

	return nil
}

// New creates a new logger instance
func New(level LogLevel, writer io.Writer, prefix string) *Logger {
	return &Logger{
		level:      level,
		fileWriter: writer,
		prefix:     prefix,
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, format, v...)
	}
}

// Warning logs a warning message
func Warning(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARNING, format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, v...)
	}
}

// Fatal logs an error message and exits
func Fatal(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, v...)
	}
	os.Exit(1)
}

// log writes a log entry
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelStrings[level]
	message := fmt.Sprintf(format, v...)

	var logLine string
	if l.prefix != "" {
		logLine = fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, levelStr, l.prefix, message)
	} else {
		logLine = fmt.Sprintf("[%s] [%s] %s\n", timestamp, levelStr, message)
	}

	if l.fileWriter != nil {
		_, _ = l.fileWriter.Write([]byte(logLine))
	} else {
		log.Print(logLine)
	}
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING", "WARN":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// GetLevel returns current log level
func GetLevel() LogLevel {
	if defaultLogger != nil {
		return defaultLogger.level
	}
	return INFO
}
