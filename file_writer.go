package timber

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"
)

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

type FileWriter struct {
	wr              *BufferedWriter
	BaseFilename    string
	currentFilename string
	mutex           *sync.RWMutex
}

// This writer has a buffer that I don't ever bother to flush, so it may take a while
// to see messages.  Filenames ending in .gz will automatically be compressed on write.
// Filename string is proccessed through the template library using the FilenameFields
// struct.
func NewFileWriter(name string) (*FileWriter, error) {
	w := &FileWriter{
		BaseFilename: name,
		mutex:        new(sync.RWMutex),
	}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *FileWriter) open() error {
	name := preprocessFilename(w.BaseFilename)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("TIMBER! Can't open %v: %v", name, err)
	}

	var output io.WriteCloser = file
	// Wrap in gz writer
	if strings.HasSuffix(name, ".gz") {
		output = &gzFileWriter{
			gzip.NewWriter(output),
			output,
		}
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.wr != nil {
		w.wr.Close()
	}
	w.currentFilename = name
	w.wr, _ = NewBufferedWriter(output)

	return nil
}

func (w *FileWriter) LogWrite(m string) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	w.wr.LogWrite(m)
}

func (w *FileWriter) Flush() error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.wr.Flush()
}

func (w *FileWriter) Rotate() error {
	return w.open()
}

func (w *FileWriter) Close() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.wr.Close()
	w.wr = nil
}

type gzFileWriter struct {
	*gzip.Writer // the compressor
	file         io.WriteCloser
}

func (w *gzFileWriter) Close() error {
	w.Writer.Close()
	return w.file.Close()
}
