// Package kit Create at 2020-12-01 11:11
package pocket

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

// Level type
type Level uint32

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

type Logback struct {
	logger *log.Logger
	fields *Fields
	level  Level
}

type Fields map[string]interface{}

type Api interface {
	Info(v ...interface{})
	Debug(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Trace(v ...interface{})
}

const defaultFormat = "| %s | %s:%d |- %s"
const defaultFlag = log.Ldate | log.Ltime

var DefaultLogger = &Logback{
	logger: log.New(os.Stdout, "", defaultFlag),
	fields: nil,
	level:  InfoLevel,
}

func (logback *Logback) SetLevel(level Level) {
	logback.level = level
}

func (logback *Logback) SetOutput(w io.Writer) {
	logback.logger.SetOutput(w)
}

func (logback *Logback) WithFields(fields Fields) *Logback {
	logback.fields = &fields
	return logback
}

func (logback *Logback) Info(v ...interface{}) {
	if logback.level < InfoLevel {
		return
	}
	logback.output("INFO", v...)
}

func (logback *Logback) Debug(v ...interface{}) {
	if logback.level < DebugLevel {
		return
	}
	logback.output("DEBUG", v...)
}

func (logback *Logback) Warn(v ...interface{}) {
	if logback.level < WarnLevel {
		return
	}
	logback.output("WARN", v...)
}

func (logback *Logback) Trace(v ...interface{}) {
	if logback.level < TraceLevel {
		return
	}
	logback.output("TRACE", v...)
}

func (logback *Logback) Error(v ...interface{}) {
	if logback.level < ErrorLevel {
		return
	}
	logback.output("ERROR", v...)
}

func (logback *Logback) output(level string, v ...interface{}) {
	format := defaultFormat
	file, line := getCaller(3)
	fields := ""
	str := ""
	if nil != logback.fields {
		format = "| %s | %s:%d | %s- %s"
		for k, v := range *(logback.fields) {
			fields += k + ":" + getLogbackValue(v) + " | "
		}
		fields = fields[:len(fields)-1]
		str = fmt.Sprintf(format, level, file, line, fields, fmt.Sprintln(v...))
	} else {
		str = fmt.Sprintf(format, level, file, line, fmt.Sprintln(v...))
	}

	logback.logger.Print(str)
}

func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line
}

func getLogbackValue(v interface{}) string {
	if nil == v {
		return ""
	}
	switch v.(type) {
	case string:
		return fmt.Sprintf("%s", v)
	case int, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case *string:
		if nil == v.(*string) {
			return ""
		}
		return fmt.Sprintf("%s", *(v.(*string)))
	case *int:
		if nil == v.(*int) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*int)))
	case *int64:
		if nil == v.(*int64) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*int64)))
	case *uint:
		if nil == v.(*uint) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*uint)))
	case *uint8:
		if nil == v.(*uint8) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*uint8)))
	case *uint16:
		if nil == v.(*uint16) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*uint16)))
	case *uint32:
		if nil == v.(*uint32) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*uint32)))
	case *uint64:
		if nil == v.(*uint64) {
			return ""
		}
		return fmt.Sprintf("%d", *(v.(*uint64)))
	default:
		return ""
	}
}
