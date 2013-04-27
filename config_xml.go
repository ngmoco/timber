package timber

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"reflect"
)

// Granulars are overriding levels that can be either
// package paths or package path + function name
type XMLGranular struct {
	Level string `xml:"level"`
	Path  string `xml:"path"`
}

// match the log4go structure so i don't have to change my configs
type XMLProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}
type XMLFilter struct {
	XMLName    xml.Name      `xml:"filter"`
	Enabled    bool          `xml:"enabled,attr"`
	Tag        string        `xml:"tag"`
	Type       string        `xml:"type"`
	Level      string        `xml:"level"`
	Format     XMLProperty   `xml:"format"`
	Properties []XMLProperty `xml:"property"`
	Granulars  []XMLGranular    `xml:"granular"`
}

type XMLConfig struct {
	XMLName xml.Name    `xml:"logging"`
	Filters []XMLFilter `xml:"filter"`
}

// Loads the configuration from an XML file (as you were probably expecting)
func (t *Timber) LoadXMLConfig(filename string) {
	if len(filename) <= 0 {
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't load xml config file: %s %v", filename, err))
	}
	defer file.Close()

	config := XMLConfig{}
	err = xml.NewDecoder(file).Decode(&config)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't parse xml config file: %s %v", filename, err))
	}

	for _, filter := range config.Filters {
		if !filter.Enabled {
			continue
		}
		level := getLevel(filter.Level)
		formatter := getXMLFormatter(filter)
		granulars := make(map[string]Level)
		for _, granular := range filter.Granulars {
			granulars[granular.Path] = getLevel(granular.Level)
		}
		configLogger := ConfigLogger{Level: level, Formatter: formatter, Granulars: granulars}

		switch filter.Type {
		case "console":
			configLogger.LogWriter = new(ConsoleWriter)
		case "socket":
			configLogger.LogWriter = getXMLSocketWriter(filter)
		case "file":
			configLogger.LogWriter = getXMLFileWriter(filter)
		default:
			log.Printf("TIMBER! Warning unrecognized filter in config file: %v\n", filter.Tag)
			continue
		}

		t.AddLogger(configLogger)
	}
}

func getXMLFormatter(filter XMLFilter) LogFormatter {
	format := ""
	property := XMLProperty{}

	// If format field is set then use it's value, otherwise
	// attempt to get the format field from the filters properties
	if !reflect.DeepEqual(filter.Format, property) {
		format = filter.Format.Value
	} else {
		for _, prop := range filter.Properties {
			if prop.Name == "format" {
				format = prop.Value
			}
		}
	}

	// If empty format set the default as just the message
	if format == "" {
		format = "%M"
	}
	return NewPatFormatter(format)
}

func getXMLSocketWriter(filter XMLFilter) LogWriter {
	var protocol, endpoint string

	for _, property := range filter.Properties {
		if property.Name == "protocol" {
			protocol = property.Value
		} else if property.Name == "endpoint" {
			endpoint = property.Value
		}
	}

	if protocol == "" || endpoint == "" {
		panic("TIMBER! Missing protocol or endpoint for socket log writer")
	}
	return NewSocketWriter(protocol, endpoint)
}

func getXMLFileWriter(filter XMLFilter) LogWriter {
	filename := ""

	for _, property := range filter.Properties {
		if property.Name == "filename" {
			filename = property.Value
		}
	}
	if filename == "" {
		panic("TIMBER! Missing filename for file log writer")
	}
	return NewFileWriter(filename)
}
