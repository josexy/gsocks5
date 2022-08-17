package logx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

type LoggerLevel int

const (
	LevelDebug LoggerLevel = 1 << iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type LoggerFlag int

const (
	FlagPrefix LoggerFlag = 1 << iota
	FlagDatetime
	FlagTime
	FlagLineNumber
	FlagFunction

	StdLoggerFlags = FlagPrefix | FlagDatetime | FlagLineNumber | FlagFunction
)

var (
	DisableLogger bool
	DisableColor  bool
	DisableDebug  bool
	StdOutput     = color.Output
	StdLoggerX    = NewLoggerX(StdOutput, StdLoggerFlags)

	strColorLoggerLevel = map[LoggerLevel]string{
		LevelDebug: HiCyan("DEBUG"),
		LevelInfo:  Green("INFO"),
		LevelWarn:  Yellow("WARN"),
		LevelError: Red("ERROR"),
		LevelFatal: HiRed("FATAL"),
	}

	strLoggerLevel = map[LoggerLevel]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
	}
)

type LoggerX struct {
	mu        sync.Mutex
	flag      LoggerFlag
	buf       []byte
	out       io.Writer
	isDiscard int32
}

func NewLoggerX(w io.Writer, flag LoggerFlag) *LoggerX {
	l := &LoggerX{
		out:  w,
		flag: flag,
	}
	if w == io.Discard {
		l.isDiscard = 1
	}
	return l
}

func (l *LoggerX) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	isDiscard := int32(0)
	if w == io.Discard {
		isDiscard = 1
	}
	atomic.StoreInt32(&l.isDiscard, isDiscard)
}

func (l *LoggerX) Debug(format string, args ...any) {
	if DisableDebug {
		return
	}
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.output(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *LoggerX) Info(format string, args ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.output(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *LoggerX) Warn(format string, args ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.output(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *LoggerX) Error(format string, args ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.output(LevelError, fmt.Sprintf(format, args...))
}

func (l *LoggerX) Fatal(format string, args ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.output(LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *LoggerX) output(level LoggerLevel, s string) {
	if DisableLogger {
		return
	}
	if DisableColor {
		color.NoColor = true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	l.buf = l.buf[:0]
	l.format(level, &l.buf, now)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, _ = l.out.Write(l.buf)
}

// format example: [INFO] [1919/08/10 11:45:14] [main.go:10#funcname] this is a log message
func (l *LoggerX) format(level LoggerLevel, buf *[]byte, now time.Time) {
	if l.flag&FlagPrefix != 0 {
		*buf = append(*buf, '[')
		mp := strColorLoggerLevel
		if DisableColor {
			mp = strLoggerLevel
		}
		*buf = append(*buf, mp[level]...)
		*buf = append(*buf, ']', ' ')
	}
	if l.flag&(FlagDatetime|FlagTime) != 0 {
		*buf = append(*buf, '[')
		var timeFormat string
		if l.flag&FlagDatetime != 0 {
			timeFormat = "2006/01/02 15:04:05"
		} else if l.flag&FlagTime != 0 {
			timeFormat = "15:04:05"
		}
		*buf = append(*buf, Blue(now.Format(timeFormat))...)
		*buf = append(*buf, ']', ' ')
	}
	if l.flag&(FlagLineNumber|FlagFunction) != 0 {
		*buf = append(*buf, '[')
		pc, file, line, _ := runtime.Caller(3)
		if l.flag&FlagLineNumber != 0 {
			*buf = append(*buf, Green(filepath.Base(file))...)
			*buf = append(*buf, ':')
			*buf = append(*buf, Red(strconv.Itoa(line))...)
		}
		if l.flag&FlagFunction != 0 {
			ls := strings.Split(runtime.FuncForPC(pc).Name(), ".")
			*buf = append(*buf, '#')
			*buf = append(*buf, Yellow(ls[len(ls)-1])...)
		}
		*buf = append(*buf, ']', ' ')
	}
}

func Debug(format string, args ...any) {
	if DisableDebug {
		return
	}
	if atomic.LoadInt32(&StdLoggerX.isDiscard) != 0 {
		return
	}
	StdLoggerX.output(LevelDebug, fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	if atomic.LoadInt32(&StdLoggerX.isDiscard) != 0 {
		return
	}
	StdLoggerX.output(LevelInfo, fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	if atomic.LoadInt32(&StdLoggerX.isDiscard) != 0 {
		return
	}
	StdLoggerX.output(LevelWarn, fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	if atomic.LoadInt32(&StdLoggerX.isDiscard) != 0 {
		return
	}
	StdLoggerX.output(LevelError, fmt.Sprintf(format, args...))
}

func ErrorBy(err error) {
	if err == nil {
		return
	}
	if atomic.LoadInt32(&StdLoggerX.isDiscard) != 0 {
		return
	}
	StdLoggerX.output(LevelError, fmt.Sprintf("%s", err.Error()))
}

func Fatal(format string, args ...any) {
	if atomic.LoadInt32(&StdLoggerX.isDiscard) != 0 {
		return
	}
	StdLoggerX.output(LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}
