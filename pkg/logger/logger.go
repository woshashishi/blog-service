package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime"
	"time"
)

type Level int8
type Fields map[string]interface{}

const (
	// iota 是 Go 的常量计数器，从0开始，每行自动+1
	LevelDebug Level = iota // LevelDebug = 0
	LevelInfo               // LevelInfo = 1（自动递增）
	LevelWarn               // LevelWarn = 2
	LevelError              // LevelError = 3
	LevelFatal              // LevelFatal = 4
	LevelPanic              // LevelPanic = 5
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	}
	// 兜底：若传入无效的 Level 值，返回空字符串
	return ""
}

type Logger struct {
	newLogger *log.Logger // log便准库里的logger方法
	ctx       context.Context
	fields    Fields
	callers   []string
}

func NewLogger(w io.Writer, prefix string, flag int) *Logger {
	l := log.New(w, prefix, flag)
	return &Logger{
		newLogger: l,
	}
}
func (l *Logger) clone() *Logger {
	nl := *l
	return &nl
}

func (l *Logger) WithFields(fields Fields) *Logger {
	ll := l.clone()
	if ll.fields == nil {
		ll.fields = make(Fields)
	}
	for k, v := range fields {
		ll.fields[k] = v
	}
	return ll
}
func (l *Logger) WithContext(c context.Context) *Logger {
	ll := l.clone()
	ll.ctx = c
	return ll
}
func (l *Logger) WithCaller(skip int) *Logger {
	ll := l.clone()
	// 获取调用栈信息：skip表示跳过的调用层级（避免记录日志库内部的调用）
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		f := runtime.FuncForPC(pc) // 获取调用的函数名
		// 格式化调用信息：文件路径: 行号 函数名
		ll.callers = []string{fmt.Sprintf("%s: %d %s", file, line, f.Name())}
	}
	return ll
}
func (l *Logger) WithCallersFrames() *Logger {
	maxCallerDepth := 25 // 最大调用栈深度
	minCallerDepth := 1  // 最小跳过层数
	callers := []string{}
	pcs := make([]uintptr, maxCallerDepth)
	// 获取调用栈的程序计数器（PC）列表
	depth := runtime.Callers(minCallerDepth, pcs)
	// 解析PC为具体的调用帧（文件、行号、函数）
	frames := runtime.CallersFrames(pcs[:depth])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		callers = append(callers, fmt.Sprintf("%s: %d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	ll := l.clone()
	ll.callers = callers // 保存完整调用栈
	return ll
}
func (l *Logger) JSONFormat(level Level, message string) map[string]interface{} {
	data := make(Fields, len(l.fields)+4)
	data["level"] = level.String()
	data["time"] = time.Now().Local().UnixNano()
	data["message"] = message
	data["callers"] = l.callers
	if len(l.fields) > 0 {
		for k, v := range l.fields {
			if _, ok := data[k]; !ok {
				data[k] = v
			}
		}
	}

	return data
}

func (l *Logger) Output(level Level, message string) {
	body, _ := json.Marshal(l.JSONFormat(level, message))
	content := string(body)
	switch level {
	case LevelDebug:
		l.newLogger.Print(content)
	case LevelInfo:
		l.newLogger.Print(content)
	case LevelWarn:
		l.newLogger.Print(content)
	case LevelError:
		l.newLogger.Print(content)
	case LevelFatal:
		l.newLogger.Fatal(content)
	case LevelPanic:
		l.newLogger.Panic(content)
	}
}
func (l *Logger) Info(v ...interface{}) {
	l.Output(LevelInfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Output(LevelFatal, fmt.Sprint(v...))
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(LevelFatal, fmt.Sprintf(format, v...))
}
func (l *Logger) Errorf(v ...interface{}) {
	l.Output(LevelError, fmt.Sprint(v...))
}
