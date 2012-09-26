// This is a brand new logger implementation that matches the log4go interface and also may be
// used as a drop-in replacement for the standard logger
//
// Basic use:
//   import log "timber"
//   log.LoadConfiguration("timber.xml")
//   log.Debug("Debug message!")
//
// XML Config file:
//		<logging>
//		  <filter enabled="true">
//			<tag>stdout</tag>
//			<type>console</type>
//			<!-- level is (:?FINEST|FINE|DEBUG|TRACE|INFO|WARNING|ERROR) -->
//			<level>DEBUG</level>
//		  </filter>
//		  <filter enabled="true">
//			<tag>file</tag>
//			<type>file</type>
//			<level>FINEST</level>
//			<property name="filename">log/server.log</property>
//			<property name="format">server [%D %T] [%L] %M</property>
//		  </filter>
//		  <filter enabled="false">
//			<tag>syslog</tag>
//			<type>socket</type>
//			<level>FINEST</level>
//			<property name="protocol">unixgram</property>
//			<property name="endpoint">/dev/log</property>
//		    <format name="pattern">%L %M</property>
//		  </filter>
//		</logging>
// The <tag> is ignored.
//
// To configure the pattern formatter all filters accept:
//		<format name="pattern">[%D %T] %L %M</format>
// Pattern format specifiers (not the same as log4go!):
// 		%T - Time: 17:24:05.333 HH:MM:SS.ms
// 		%t - Time: 17:24:05 HH:MM:SS
// 		%D - Date: 2011-12-25 yyyy-mm-dd
// 		%d - Date: 2011/12/25 yyyy/mm/dd
// 		%L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// 		%S - Source: full runtime.Caller line and line number
// 		%s - Short Source: just file and line number
// 		%x - Extra Short Source: just file without .go suffix
// 		%M - Message
// 		%% - Percent sign
// the string number prefixes are allowed e.g.: %10s will pad the source field to 10 spaces
// pattern defaults to %M
// Both log4go synatax of <property name="format"> and new <format name=type> are supported
// the property syntax will only ever support the pattern formatter
//
// Code Architecture:
// A MultiLogger <logging> which consists of many ConfigLoggers <filter>. ConfigLoggers have three properties:
// LogWriter <type>, Level (as a threshold) <level> and LogFormatter <format>.
//
// In practice, this means that you define ConfigLoggers with a LogWriter (where the log prints to
// eg. socket, file, stdio etc), the Level threshold, and a LogFormatter which formats the message
// before writing.  Because the LogFormatters and LogWriters are simple interfaces, it is easy to
// write your own custom implementations.
//
// Once configured, you only deal with the "Logger" interface and use the log methods in your code
//
// The motivation for this package grew from a need to make some changes to the functionality of
// log4go (which had already been integrated into a larger project).  I tried to maintain compatiblity
// with log4go for the interface and configuration.  The main issue I had with log4go was that each of
// logger types had incisistent and incompatible configuration.  I looked at contributing changes to
// log4go, but I would have needed to break existing use cases so I decided to do a rewrite from scratch.
//
package timber

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"
)

type Level int

// Log levels
const (
	NONE Level = iota // NONE to be used for standard go log impl's
	FINEST
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Default level passed to runtime.Caller by Timber, add to this if you wrap Timber in your own logging code
const DefaultFileDepth int = 3

// What gets printed for each Log level
var LevelStrings = [...]string{"", "FNST", "FINE", "DEBG", "TRAC", "INFO", "WARN", "EROR", "CRIT"}

// Full level names
var LongLevelStrings = []string{
	"NONE",
	"FINEST",
	"FINE",
	"DEBUG",
	"TRACE",
	"INFO",
	"WARNING",
	"ERROR",
	"CRITICAL",
}

// Return a given level string as the actual Level value
func getLevel(lvlString string) Level {
	for idx, str := range LongLevelStrings {
		if str == lvlString {
			return Level(idx)
		}
	}
	return Level(0)
}

// This explicitly defines the contract for a logger
// Not really useful except for documentation for
// writing an separate implementation
type Logger interface {
	// match log4go interface to drop-in replace
	Finest(arg0 interface{}, args ...interface{})
	Fine(arg0 interface{}, args ...interface{})
	Debug(arg0 interface{}, args ...interface{})
	Trace(arg0 interface{}, args ...interface{})
	Info(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{}) error
	Error(arg0 interface{}, args ...interface{}) error
	Critical(arg0 interface{}, args ...interface{}) error
	Log(lvl Level, arg0 interface{}, args ...interface{})

	// support standard log too
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
}

// Not used
type LoggerConfig interface {
	// When set, messages with level < lvl will be ignored.  It's up to the implementor to keep the contract or not
	SetLevel(lvl Level)
	// Set the formatter for the log
	SetFormatter(formatter LogFormatter)
}

// Interface required for a log writer endpoint.  It's more or less a
// io.WriteCloser with no errors allowed to be returned and string
// instead of []byte.
//
// TODO: Maybe this should just be a standard io.WriteCloser?
type LogWriter interface {
	LogWrite(msg string)
	Close()
}

// This packs up all the message data and metadata. This structure
// will be passed to the LogFormatter
type LogRecord struct {
	Level      Level
	Timestamp  int64
	SourceFile string
	SourceLine int
	Message    string
}

// Format a log message before writing
type LogFormatter interface {
	Format(rec LogRecord) string
}

// Container a single log format/destination
type ConfigLogger struct {
	LogWriter LogWriter
	// Messages with level < Level will be ignored.  It's up to the implementor to keep the contract or not
	Level     Level
	Formatter LogFormatter
}

// Allow logging to multiple places
type MultiLogger interface {
	// returns an int that identifies the logger for future calls to SetLevel and SetFormatter
	AddLogger(logger ConfigLogger) int
	// dynamically change level or format
	SetLevel(index int, lvl Level)
	SetFormatter(index int, formatter LogFormatter)
	Close()
}

//
//
//
// Implementation
//
//
//

// The Timber instance is the concrete implementation of the logger interfaces.
// New instances may be created, but usually you'll just want to use the default
// instance in Global
//
// NOTE: I don't supporting the log4go special handling of the first parameter based on type
// mainly cuz I don't think it's particularly useful (I kept passing a data string as the first
// param and expecting a Println-like output but that would always break expecting a format string)
// I also don't support the passing of the closure stuff
type Timber struct {
	writerConfigChan chan timberConfig
	recordChan       chan LogRecord
	hasLogger        bool
	// This value is passed to runtime.Caller to get the file name/line and may require
	// tweaking if you want to wrap the logger
	FileDepth int
}

type timberAction int

const (
	actionAdd timberAction = iota
	actionModify
	actionQuit
)

type timberConfig struct {
	Action timberAction // type of config action
	Index  int          // only for modify
	Cfg    ConfigLogger // used for modify or add
	Ret    chan int     // only used for add
}

// Creates a new Timber logger that is ready to be configured
// With no subsequent configuration, nothing will be logged
//
func NewTimber() *Timber {
	t := new(Timber)
	t.writerConfigChan = make(chan timberConfig)
	t.recordChan = make(chan LogRecord, 300)
	t.FileDepth = DefaultFileDepth
	go t.asyncLumberJack()
	return t
}

func (t *Timber) asyncLumberJack() {
	var loggers []ConfigLogger = make([]ConfigLogger, 0, 2)
	loopIt := true
	for loopIt {
		select {
		case rec := <-t.recordChan:
			sendToLoggers(loggers, rec)
		case cfg := <-t.writerConfigChan:
			switch cfg.Action {
			case actionAdd:
				loggers = append(loggers, cfg.Cfg)
				cfg.Ret <- (len(loggers) - 1)
			case actionModify:
			case actionQuit:
				close(t.recordChan)
				loopIt = false
				defer func() {
					cfg.Ret <- 0
				}()
			}
		} // select
	} // for
	// drain the log channel before closing
	for rec := range t.recordChan {
		sendToLoggers(loggers, rec)
	}
	closeAllWriters(loggers)
}

func sendToLoggers(loggers []ConfigLogger, rec LogRecord) {
	formatted := ""
	for _, cLog := range loggers {
		if rec.Level >= cLog.Level || rec.Level == 0 {
			if formatted == "" {
				formatted = cLog.Formatter.Format(rec)
			}
			cLog.LogWriter.LogWrite(formatted)
		}
	}
}

func closeAllWriters(cls []ConfigLogger) {
	for _, cLog := range cls {
		cLog.LogWriter.Close()
	}
}

// MultiLogger interface
func (t *Timber) AddLogger(logger ConfigLogger) int {
	tcChan := make(chan int, 1) // buffered
	tc := timberConfig{Action: actionAdd, Cfg: logger, Ret: tcChan}
	t.writerConfigChan <- tc
	return <-tcChan
}

// MultiLogger interface
func (t *Timber) Close() {
	tcChan := make(chan int)
	tc := timberConfig{Action: actionQuit, Ret: tcChan}
	t.writerConfigChan <- tc
	<-tcChan // block for closing
}

// Not yet implemented
func (t *Timber) SetLevel(index int, lvl Level) {
	// TODO
}

// Not yet implemented
func (t *Timber) SetFormatter(index int, formatter LogFormatter) {
	// TODO
}

// Logger interface
func (t *Timber) prepareAndSend(lvl Level, msg string, depth int) {
	now := time.Now().UnixNano()
	_, file, line, _ := runtime.Caller(depth)
	t.recordChan <- LogRecord{Level: lvl, Timestamp: now, SourceFile: file, SourceLine: line, Message: msg}
}

func (t *Timber) Finest(arg0 interface{}, args ...interface{}) {
	t.prepareAndSend(FINEST, fmt.Sprintf(arg0.(string), args...), t.FileDepth)
}
func (t *Timber) Fine(arg0 interface{}, args ...interface{}) {
	t.prepareAndSend(FINE, fmt.Sprintf(arg0.(string), args...), t.FileDepth)
}
func (t *Timber) Debug(arg0 interface{}, args ...interface{}) {
	t.prepareAndSend(DEBUG, fmt.Sprintf(arg0.(string), args...), t.FileDepth)
}
func (t *Timber) Trace(arg0 interface{}, args ...interface{}) {
	t.prepareAndSend(TRACE, fmt.Sprintf(arg0.(string), args...), t.FileDepth)
}
func (t *Timber) Info(arg0 interface{}, args ...interface{}) {
	t.prepareAndSend(INFO, fmt.Sprintf(arg0.(string), args...), t.FileDepth)
}
func (t *Timber) Warn(arg0 interface{}, args ...interface{}) error {
	msg := fmt.Sprintf(arg0.(string), args...)
	t.prepareAndSend(WARNING, msg, t.FileDepth)
	return errors.New(msg)
}
func (t *Timber) Error(arg0 interface{}, args ...interface{}) error {
	msg := fmt.Sprintf(arg0.(string), args...)
	t.prepareAndSend(ERROR, msg, t.FileDepth)
	return errors.New(msg)
}
func (t *Timber) Critical(arg0 interface{}, args ...interface{}) error {
	msg := fmt.Sprintf(arg0.(string), args...)
	t.prepareAndSend(CRITICAL, msg, t.FileDepth)
	return errors.New(msg)
}
func (t *Timber) Log(lvl Level, arg0 interface{}, args ...interface{}) {
	t.prepareAndSend(lvl, fmt.Sprintf(arg0.(string), args...), t.FileDepth)
}

// Print won't work well with a pattern_logger because it explicitly adds
// its own \n; so you'd have to write your own formatter to remove it
func (t *Timber) Print(v ...interface{}) {
	t.prepareAndSend(NONE, fmt.Sprint(v...), t.FileDepth)
}
func (t *Timber) Printf(format string, v ...interface{}) {
	t.prepareAndSend(NONE, fmt.Sprintf(format, v...), t.FileDepth)
}

// Println won't work well either with a pattern_logger because it explicitly adds
// its own \n; so you'd have to write your own formatter to not have 2 \n's
func (t *Timber) Println(v ...interface{}) {
	t.prepareAndSend(NONE, fmt.Sprintln(v...), t.FileDepth)
}
func (t *Timber) Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	t.prepareAndSend(NONE, msg, t.FileDepth)
	panic(msg)
}
func (t *Timber) Panicf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	t.prepareAndSend(NONE, msg, t.FileDepth)
	panic(msg)
}
func (t *Timber) Panicln(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	t.prepareAndSend(NONE, msg, t.FileDepth)
	panic(msg)
}
func (t *Timber) Fatal(v ...interface{}) {
	msg := fmt.Sprint(v...)
	t.prepareAndSend(NONE, msg, t.FileDepth)
	t.Close()
	os.Exit(1)
}
func (t *Timber) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	t.prepareAndSend(NONE, msg, t.FileDepth)
	t.Close()
	os.Exit(1)
}
func (t *Timber) Fatalln(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	t.prepareAndSend(NONE, msg, t.FileDepth)
	t.Close()
	os.Exit(1)
}

//
//
// Default Instance
//
//

// Default Timber Instance (used for all the package level function calls)
var Global = NewTimber()

// Simple wrappers for Logger interface
func Finest(arg0 interface{}, args ...interface{})         { Global.Finest(arg0, args...) }
func Fine(arg0 interface{}, args ...interface{})           { Global.Fine(arg0, args...) }
func Debug(arg0 interface{}, args ...interface{})          { Global.Debug(arg0, args...) }
func Trace(arg0 interface{}, args ...interface{})          { Global.Trace(arg0, args...) }
func Info(arg0 interface{}, args ...interface{})           { Global.Info(arg0, args...) }
func Warn(arg0 interface{}, args ...interface{}) error     { return Global.Warn(arg0, args...) }
func Error(arg0 interface{}, args ...interface{}) error    { return Global.Error(arg0, args...) }
func Critical(arg0 interface{}, args ...interface{}) error { return Global.Critical(arg0, args...) }
func Log(lvl Level, arg0 interface{}, args ...interface{}) { Global.Log(lvl, arg0, args...) }
func Print(v ...interface{})                               { Global.Print(v...) }
func Printf(format string, v ...interface{})               { Global.Printf(format, v...) }
func Println(v ...interface{})                             { Global.Println(v...) }
func Panic(v ...interface{})                               { Global.Panic(v...) }
func Panicf(format string, v ...interface{})               { Global.Panicf(format, v...) }
func Panicln(v ...interface{})                             { Global.Panicln(v...) }
func Fatal(v ...interface{})                               { Global.Fatal(v...) }
func Fatalf(format string, v ...interface{})               { Global.Fatalf(format, v...) }
func Fatalln(v ...interface{})                             { Global.Fatalln(v...) }

func AddLogger(logger ConfigLogger) int { return Global.AddLogger(logger) }
func Close()                            { Global.Close() }

func LoadConfiguration(filename string)     { Global.LoadConfig(filename) }
func LoadXMLConfiguration(filename string)  { Global.LoadXMLConfig(filename) }
func LoadJSONConfiguration(filename string) { Global.LoadJSONConfig(filename) }
