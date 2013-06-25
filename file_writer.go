package timber

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"
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

type FilenameFields struct {
	Hostname string
	Date     time.Time
	Pid      int
}

func GetFilenameFields() *FilenameFields {
	h, _ := os.Hostname()
	return &FilenameFields{
		Hostname: h,
		Date:     time.Now(),
		Pid:      os.Getpid(),
	}
}

func preprocessFilename(name string) string {
	t := template.Must(template.New("filename").Parse(name))
	buf := new(bytes.Buffer)
	t.Execute(buf, GetFilenameFields())
	return buf.String()
}

// This writer has a buffer that I don't ever bother to flush, so it may take a while
// to see messages.  Filenames ending in .gz will automatically be compressed on write.
// Filename string is proccessed through the template library using the FilenameFields
// struct.
func NewFileWriter(name string) (LogWriter, error) {
	name = preprocessFilename(name)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("TIMBER! Can't open %v: %v", name, err)
	}
	var output io.WriteCloser = file
	// Wrap in gz writer
	if strings.HasSuffix(name, ".gz") {
		output = &gzFileWriter{
			gzip.NewWriter(output),
			output,
		}
	}
	return NewBufferedWriter(output)
}

type gzFileWriter struct {
	*gzip.Writer // the compressor
	file         io.WriteCloser
}

func (w *gzFileWriter) Close() error {
	w.Writer.Close()
	return w.file.Close()
}
