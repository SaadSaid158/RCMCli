package logger

import (
	"fmt"
	"os"
)

// Logger provides simple logging with [*], [+], [-] format
type Logger struct {
	verbose bool
}

// New creates a new logger
func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// Close is a no-op for compatibility
func (l *Logger) Close() error {
	return nil
}

// Info logs an info message with [*]
func (l *Logger) Info(msg string, args ...interface{}) {
	fmt.Printf("[*] "+msg+"\n", args...)
}

// Success logs a success message with [+]
func (l *Logger) Success(msg string, args ...interface{}) {
	fmt.Printf("[+] "+msg+"\n", args...)
}

// Error logs an error message with [-]
func (l *Logger) Error(msg string, args ...interface{}) {
	fmt.Printf("[-] "+msg+"\n", args...)
}

// Debug logs a debug message (only if verbose)
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.verbose {
		fmt.Printf("[*] "+msg+"\n", args...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	fmt.Printf("[-] "+msg+"\n", args...)
	os.Exit(1)
}
