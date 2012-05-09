package timber

import (
	"fmt"
	"os"
)

/* unbuffered impl
type FileWriter struct {
	file io.WriteCloser
}
func NewFileWriter(name string) (*FileWriter) {
	fw := new(FileWriter)
	file, err := os.OpenFile(name, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0666)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't open: %v", name))
	}
	fw.file = file
	return fw
}

func (fw *FileWriter) LogWrite(msg string) {
	fw.file.Write([]byte(msg))
}

func (fw *FileWriter) Close() {
	fw.file.Close()
}
*/

// This writer has a buffer that I don't ever bother to flush, so it may take a while
// to see messages
func NewFileWriter(name string) LogWriter {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't open: %v", name))
	}
	return NewBufferedWriter(file)
}
