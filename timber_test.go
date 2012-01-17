package timber

import (
	"testing"
)

func TestConsole(t *testing.T) {
	log := NewTimber()
	console := new(ConsoleWriter)
	formatter := NewPatFormatter("[%D %T] [%L] %-10x %M")
	idx := log.AddLogger(ConfigLogger{LogWriter: console,
		Level:     DEBUG,
		Formatter: formatter})
	log.Error("what error? %v", idx)
	log.Close()
}

func TestFile(t *testing.T) {
	log := NewTimber()
	writer := NewFileWriter("test.log")
	formatter := NewPatFormatter("[%D %T] [%L] %-10x %M")
	idx := log.AddLogger(ConfigLogger{LogWriter: writer,
		Level:     FINEST,
		Formatter: formatter})
	log.Error("what error? %v", idx)
	log.Warn("I'm waringing you!")
	log.Info("FYI")
	log.Fine("you soo fine!")
	log.Close()
}

func TestXmlConfig(t *testing.T) {
	log := NewTimber()
	log.LoadXMLConfig("timber.xml")
	log.Info("Message to XML loggers")
	log.Close()
}

func TestDefaultLogger(t *testing.T) {
	console := new(ConsoleWriter)
	formatter := NewPatFormatter("%DT%T %L %-10x %M")
	AddLogger(ConfigLogger{LogWriter: console,
		Level:     DEBUG,
		Formatter: formatter})
	Warn("Some sweet default logging")
	Close()
}
