package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	logLevel = INFO
	logger   = log.New(os.Stdout, "", 0)
)

func init() {
	// Set log level from environment
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = DEBUG
		case "INFO":
			logLevel = INFO
		case "WARN":
			logLevel = WARN
		case "ERROR":
			logLevel = ERROR
		}
	}
}

func formatMessage(level string, msg string, fields ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	if ok {
		file = file[strings.LastIndex(file, "/")+1:]
	} else {
		file = "unknown"
		line = 0
	}

	// Format fields as key=value pairs
	var fieldStr strings.Builder
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			fieldStr.WriteString(fmt.Sprintf(" %v=%v", fields[i], fields[i+1]))
		}
	}

	return fmt.Sprintf("[%s] %s %s:%d %s%s", level, timestamp, file, line, msg, fieldStr.String())
}

func Debug(msg string, fields ...interface{}) {
	if logLevel <= DEBUG {
		logger.Println(formatMessage("DEBUG", msg, fields...))
	}
}

func Info(msg string, fields ...interface{}) {
	if logLevel <= INFO {
		logger.Println(formatMessage("INFO", msg, fields...))
	}
}

func Warn(msg string, fields ...interface{}) {
	if logLevel <= WARN {
		logger.Println(formatMessage("WARN", msg, fields...))
	}
}

func Error(msg string, fields ...interface{}) {
	if logLevel <= ERROR {
		logger.Println(formatMessage("ERROR", msg, fields...))
	}
}

func Fatal(msg string, fields ...interface{}) {
	logger.Println(formatMessage("FATAL", msg, fields...))
	os.Exit(1)
}

// SetLevel sets the logging level
func SetLevel(level Level) {
	logLevel = level
}

// SetOutput sets the output destination for the logger
func SetOutput(output *os.File) {
	logger.SetOutput(output)
}