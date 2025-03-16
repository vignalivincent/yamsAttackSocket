package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

type Logger struct {
	level     string
	color     string
	writer    io.Writer
	skipDepth int
}

func (l *Logger) CustomLog(format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(l.skipDepth)
	fileInfo := ""
	if ok {
		fileInfo = fmt.Sprintf("# %s:%d", filepath.Base(file), line)
	}

	timestamp := time.Now().Format("15:04")
	message := fmt.Sprintf(format, args...)
	
	// Calculate how much padding is needed
	// Assuming terminal width of 120 characters and file info length
	terminalWidth := 120
	prefixLength := len(timestamp) + len(l.level) + 6 // 6 accounts for brackets, colons, etc.
	messageLength := len(message)
	fileInfoLength := len(fileInfo)
	
	// Calculate padding needed to push file info to the right
	padding := terminalWidth - prefixLength - messageLength - fileInfoLength
	if padding < 4 {
		padding = 4 // Minimum 4 spaces of separation
	}
	
	// Create the padding string
	paddingStr := strings.Repeat(" ", padding)
	
	// Format and output the log entry
	logEntry := fmt.Sprintf("[%s] %s%s%s : %s%s%s\n", 
		timestamp, 
		l.color, l.level, colorReset,
		message,
		paddingStr,
		fileInfo)
	
	fmt.Fprint(l.writer, logEntry)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.CustomLog(format, args...)
}

func (l *Logger) Println(args ...interface{}) {
	l.CustomLog("%s", fmt.Sprint(args...))
}

var (
	Debug = &Logger{level: "DEBUG", color: colorBlue, writer: os.Stdout, skipDepth: 2}
	Info = &Logger{level: "INFO", color: colorGreen, writer: os.Stdout, skipDepth: 2}
	Warn = &Logger{level: "WARNING", color: colorYellow, writer: os.Stdout, skipDepth: 2}
	Error = &Logger{level: "ERROR", color: colorRed, writer: os.Stderr, skipDepth: 2}
	System = &Logger{level: "SYSTEM", color: colorCyan, writer: os.Stdout, skipDepth: 2}
)
