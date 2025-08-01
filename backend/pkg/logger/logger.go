package logger

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
)

func Init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(v ...interface{}) {
	InfoLogger.Println(v...)
}

func Error(v ...interface{}) {
	ErrorLogger.Println(v...)
}

func Warn(v ...interface{}) {
	WarnLogger.Println(v...)
}