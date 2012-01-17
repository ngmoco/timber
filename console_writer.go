package timber

import (
	"log"
)

func init() {
	log.SetFlags(0)
}

// This uses the standard go logger to write the messages
type ConsoleWriter func(string)

func (c ConsoleWriter) LogWrite(msg string) {
	log.Print(msg)
}

func (c ConsoleWriter) Close() {
	// Nothing
}
