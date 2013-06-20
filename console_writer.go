package timber

import (
	"fmt"
	"os"
)

// This uses the standard go logger to write the messages
type ConsoleWriter func(string)

func (c ConsoleWriter) LogWrite(msg string) {
	fmt.Fprint(os.Stderr, msg)
}

func (c ConsoleWriter) Close() {
	// Nothing
}
