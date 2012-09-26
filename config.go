package timber

import (
	"log"
	"path"
)

func (t *Timber) LoadConfig(filename string) {
	if len(filename) <= 0 {
		return
	}
	ext := path.Ext(filename)
	ext = ext[1:]

	switch ext {
	case "xml":
		t.LoadXMLConfig(filename)
		break
	case "json":
		t.LoadJSONConfig(filename)
		break
	default:
		log.Printf("TIMBER! Unknown config file type %v, only XML and JSON are supported types\n", ext)
	}
}
