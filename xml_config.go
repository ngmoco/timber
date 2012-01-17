package timber

import (
	"xml"
	"fmt"
	"os"
	"reflect"
	"log"
)

// These levels match log4go configuration
var LongLevelStrings = []string{"NONE", "FINEST", "FINE", "DEBUG", "TRACE", "INFO", "WARNING", "ERROR", "CRITICAL"}

// match the log4go structure so i don't have to change my configs
type xmlProperty struct {
	Name  string `xml:"attr"`
	Value string `xml:"chardata"`
}
type xmlFilter struct {
	XMLName  xml.Name `xml:"filter"`
	Tag      string
	Enabled  bool `xml:"attr"`
	Type     string
	Level    string
	Format   xmlProperty
	Property []xmlProperty
}

type xmlConfig struct {
	XMLName xml.Name `xml:"logging"`
	Filter  []xmlFilter
}

// Loads the configuration from an XML file (as you were probably expecting)
func (tim *Timber) LoadXMLConfig(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't load xml config file: %s %v", fileName, err))
	}

	val := xmlConfig{}
	err = xml.Unmarshal(file, &val)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't parse xml config file: %s %v", fileName, err))
	}

	for _, filter := range val.Filter {
		if filter.Enabled {
			level := getLevel(filter.Level)
			formatter := getFormatter(filter)

			switch filter.Type {
			case "console":
				tim.AddLogger(ConfigLogger{LogWriter: new(ConsoleWriter), Level: level, Formatter: formatter})
			case "socket":
				tim.AddLogger(ConfigLogger{LogWriter: getSocketWriter(filter), Level: level, Formatter: formatter})
			case "file":
				tim.AddLogger(ConfigLogger{LogWriter: getFileWriter(filter), Level: level, Formatter: formatter})
			default:
				log.Printf("TIMBER! Warning unrecognized filter in config file: %v\n", filter)
			}
		}
	}

}

func getLevel(lvlString string) level {
	for idx, str := range LongLevelStrings {
		if str == lvlString {
			return level(idx)
		}
	}
	return level(0)
}

func getFormatter(filter xmlFilter) LogFormatter {
	var format string
	// try to get the new format tag first, then fall back to the generic property one
	val := xmlProperty{}
	if !reflect.DeepEqual(filter.Format, val) { // not equal to the empty value
		format = filter.Format.Value
	} else {
		props := filter.Property
		for _, prop := range props {
			if prop.Name == `format` {
				format = prop.Value
			}
		}
	}
	if format == "" {
		format = "%M" // just the message by default
	}
	return NewPatFormatter(format)
}

func getSocketWriter(filter xmlFilter) LogWriter {
	var protocol, endpoint string
	for _, prop := range filter.Property {
		if prop.Name == "protocol" {
			protocol = prop.Value
		} else if prop.Name == "endpoint" {
			endpoint = prop.Value
		}
	}
	if protocol == "" || endpoint == "" {
		panic("TIMBER! Missing protocol or endpoint for socket log writer")
	}
	return NewSocketWriter(protocol, endpoint)
}

func getFileWriter(filter xmlFilter) LogWriter {
	var filename string
	for _, prop := range filter.Property {
		if prop.Name == "filename" {
			filename = prop.Value
		}
	}
	if filename == "" {
		panic("TIMBER! Missing filename for file log writer")
	}
	return NewFileWriter(filename)
}
