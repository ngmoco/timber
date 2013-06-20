package timber

import (
	"fmt"
	"log/syslog"
	"os"
	"time"
)

// Mapping from the timber levels to the syslog severity
// If you override this, make sure all the entries are in the map
// since the syslog.Priority zero value will cause a message at 
// the Emergency severity
var DefaultSeverityMap = map[Level]syslog.Priority{
	NONE:     syslog.LOG_INFO,
	FINEST:   syslog.LOG_DEBUG,
	FINE:     syslog.LOG_DEBUG,
	DEBUG:    syslog.LOG_DEBUG,
	TRACE:    syslog.LOG_INFO,
	INFO:     syslog.LOG_INFO,
	WARNING:  syslog.LOG_WARNING,
	ERROR:    syslog.LOG_ERR,
	CRITICAL: syslog.LOG_CRIT,
}

// Syslog formatter wraps a PatFormatter but adds the 
// syslog protocol format to the message.  
// Defaults:
// Facility: syslog.LOG_USER (1 << 3 for pre-go1.1 compatibility)
// Hostname: os.Hostname()
// Tag: os.Args[0]
type SyslogFormatter struct {
	pf          *PatFormatter
	pid         int
	Hostname    string
	Tag         string
	Facility    syslog.Priority
	SeverityMap map[Level]syslog.Priority
}

func NewSyslogFormatter(format string) *SyslogFormatter {
	hostname, _ := os.Hostname()
	return &SyslogFormatter{NewPatFormatter(format), os.Getpid(), hostname, os.Args[0], syslog.Priority(1 << 3), DefaultSeverityMap}
}

func (sf *SyslogFormatter) Format(rec *LogRecord) string {
	msg := sf.pf.Format(rec)
	return fmt.Sprintf("<%d>%s %s %s[%d]: %s",
		sf.Facility|sf.SeverityMap[rec.Level],
		rec.Timestamp.Format(time.RFC3339),
		sf.Hostname,
		sf.Tag,
		sf.pid,
		msg)
}

