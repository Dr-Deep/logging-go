package logging

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// Logging Level
type Level int

const (
	defaultLogLevel = LogDebug

	// LogDebug represents debug messages.
	LogDebug Level = iota

	// LogInfo represents informational messages.
	LogInfo

	// LogError represents errors.
	LogError

	// LogFatal represents fatal errors.
	LogFatal

	unknownFuncName = "???"
)

var (
	// logLevelDisplayNames gives the display name to use for our log levels.
	logLevelDisplayNames = map[Level]string{
		LogDebug: "debug",
		LogInfo:  "info",
		LogError: "error",
		LogFatal: "fatal",
	}
)

type Logger struct {
	sync.Mutex
	File   *os.File
	Writer *bufio.Writer
	Level  Level
}

func NewLogger(file *os.File) (logger *Logger) {
	if file == nil {
		logger = &Logger{
			File:   file,
			Level:  defaultLogLevel,
			Writer: bufio.NewWriter(&bytes.Buffer{}),
		}
	} else {
		logger = &Logger{
			File:   file,
			Level:  defaultLogLevel,
			Writer: bufio.NewWriter(file),
		}
	}

	return logger
}

func (logger *Logger) Close() error {
	flushErr := logger.Writer.Flush()
	closeErr := logger.File.Close()
	if flushErr != nil {
		return flushErr
	}

	return closeErr
}

func (logger *Logger) Log(level Level, logType string, messageParts ...string) {
	// check if we log on the right level
	if level < logger.Level {
		return
	}

	// assemble our log line
	var rawBuf bytes.Buffer
	fmt.Fprintf(
		&rawBuf,
		"[%s]-[%s]-[%s] ", time.Now().Format("2006-01-02T15:04:05.000Z"), logLevelDisplayNames[level], logType,
	)
	for i, p := range messageParts {
		rawBuf.WriteString(p)

		if i != len(messageParts)-1 {
			rawBuf.WriteString(" : ")
		}
	}
	rawBuf.WriteRune('\n')

	// output line
	logger.Lock()
	if _, err := logger.Writer.Write(rawBuf.Bytes()); err != nil {
		panic(err)
	}
	if err := logger.Writer.Flush(); err != nil {
		panic(err)
	}
	logger.Unlock()
}

func (logger *Logger) Debug(messageParts ...string) {
	pc, _, _, ok := runtime.Caller(1)
	funcname := runtime.FuncForPC(pc).Name()
	if !ok {
		funcname = unknownFuncName
	}

	logger.Log(LogDebug, funcname, messageParts...)
}

func (logger *Logger) Info(messageParts ...string) {
	pc, _, _, ok := runtime.Caller(1)
	funcname := runtime.FuncForPC(pc).Name()
	if !ok {
		funcname = unknownFuncName
	}

	logger.Log(LogInfo, funcname, messageParts...)
}

func (logger *Logger) Error(messageParts ...string) {
	pc, _, _, ok := runtime.Caller(1)
	funcname := runtime.FuncForPC(pc).Name()
	if !ok {
		funcname = unknownFuncName
	}

	logger.Log(LogError, funcname, messageParts...)
}

func (logger *Logger) Fatal(messageParts ...string) {
	pc, _, _, ok := runtime.Caller(1)
	funcname := runtime.FuncForPC(pc).Name()
	if !ok {
		funcname = unknownFuncName
	}

	logger.Log(LogFatal, funcname, messageParts...)
	os.Exit(1)
}
