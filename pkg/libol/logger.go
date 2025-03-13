package libol

import (
	"container/list"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

const (
	PRINT = 01
	LOG   = 05
	STACK = 06
	DEBUG = 10
	FLOW  = 11
	CMD   = 15
	EVENT = 16
	INFO  = 20
	WARN  = 30
	ERROR = 40
	FATAL = 99
)

type Message struct {
	Level   string `json:"level"`
	Date    string `json:"date"`
	Message string `json:"message"`
	Module  string `json:"module"`
}

var levels = map[int]string{
	PRINT: "PRINT",
	LOG:   "LOG",
	DEBUG: "DEBUG",
	STACK: "STACK",
	FLOW:  "FLOW",
	CMD:   "CMD",
	EVENT: "EVENT",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type logger struct {
	Level    int
	FileName string
	FileLog  *log.Logger
	Lock     sync.Mutex
	Errors   *list.List
}

func (l *logger) Write(level int, format string, v ...interface{}) {
	str, ok := levels[level]
	if !ok {
		str = "NULL"
	}
	if level >= l.Level {
		log.Printf(fmt.Sprintf("%s|%s", str, format), v...)
	}
	if level >= INFO {
		l.Save(str, format, v...)
	}
}

func (l *logger) Save(level string, format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	now := time.Now()
	if l.FileLog != nil {
		l.FileLog.Printf("%s|%s\n", level, m)
	}
	l.Lock.Lock()
	defer l.Lock.Unlock()
	if l.Errors.Len() >= 1024 {
		if e := l.Errors.Back(); e != nil {
			l.Errors.Remove(e)
		}
	}
	ele := &Message{
		Level:   level,
		Date:    now.Format(time.RFC3339),
		Message: m,
	}
	l.Errors.PushBack(ele)
}

func (l *logger) List() <-chan *Message {
	c := make(chan *Message, 128)
	go func() {
		l.Lock.Lock()
		defer l.Lock.Unlock()
		for ele := l.Errors.Back(); ele != nil; ele = ele.Prev() {
			c <- ele.Value.(*Message)
		}
		c <- nil // Finish channel by nil.
	}()
	return c
}

var Logger = &logger{
	Level:    INFO,
	FileName: ".log.error",
	Errors:   list.New(),
}

func SetLogger(file string, level int) {
	Logger.Level = level
	if file == "" || Logger.FileName == file {
		return
	}
	Logger.FileName = file
	fp, err := OpenWrite(file)
	if err == nil {
		Logger.FileLog = log.New(fp, "", log.LstdFlags)
	} else {
		Warn("Logger.Init: %s", err)
	}
}

func SetLevel(level int) {
	Logger.Level = level
}

type SubLogger struct {
	*logger
	Prefix string
}

func NewSubLogger(prefix string) *SubLogger {
	return &SubLogger{
		logger: Logger,
		Prefix: prefix,
	}
}

var rLogger = NewSubLogger("root")

func HasLog(level int) bool {
	return rLogger.Has(level)
}

func Catch(name string) {
	if err := recover(); err != nil {
		Fatal("%s|PANIC >>> %s <<<", name, err)
		Fatal("%s|STACK >>> %s <<<", name, debug.Stack())
	}
}

func Print(format string, v ...interface{}) {
	rLogger.Print(format, v...)
}

func Log(format string, v ...interface{}) {
	rLogger.Log(format, v...)
}

func Stack(format string, v ...interface{}) {
	rLogger.Stack(format, v...)
}

func Debug(format string, v ...interface{}) {
	rLogger.Debug(format, v...)
}

func Cmd(format string, v ...interface{}) {
	rLogger.Cmd(format, v...)
}

func Info(format string, v ...interface{}) {
	rLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	rLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	rLogger.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	rLogger.Fatal(format, v...)
}

func (s *SubLogger) Has(level int) bool {
	if level >= s.Level {
		return true
	}
	return false
}

func (s *SubLogger) Fmt(format string) string {
	return s.Prefix + "|" + format
}

func (s *SubLogger) Print(format string, v ...interface{}) {
	s.logger.Write(PRINT, s.Fmt(format), v...)
}

func (s *SubLogger) Printf(format string, v ...interface{}) {
	s.logger.Write(PRINT, s.Fmt(format), v...)
}

func (s *SubLogger) Log(format string, v ...interface{}) {
	s.logger.Write(LOG, s.Fmt(format), v...)
}

func (s *SubLogger) Stack(format string, v ...interface{}) {
	s.logger.Write(STACK, s.Fmt(format), v...)
}

func (s *SubLogger) Debug(format string, v ...interface{}) {
	s.logger.Write(DEBUG, s.Fmt(format), v...)
}

func (s *SubLogger) Flow(format string, v ...interface{}) {
	s.logger.Write(FLOW, s.Fmt(format), v...)
}

func (s *SubLogger) Cmd(format string, v ...interface{}) {
	s.logger.Write(CMD, s.Fmt(format), v...)
}

func (s *SubLogger) Event(format string, v ...interface{}) {
	s.logger.Write(EVENT, s.Fmt(format), v...)
}

func (s *SubLogger) Info(format string, v ...interface{}) {
	s.logger.Write(INFO, s.Fmt(format), v...)
}

func (s *SubLogger) Warn(format string, v ...interface{}) {
	s.logger.Write(WARN, s.Fmt(format), v...)
}

func (s *SubLogger) Error(format string, v ...interface{}) {
	s.logger.Write(ERROR, s.Fmt(format), v...)
}

func (s *SubLogger) Fatal(format string, v ...interface{}) {
	s.logger.Write(FATAL, s.Fmt(format), v...)
}

func init() {
	log.SetFlags(log.LstdFlags)
}
