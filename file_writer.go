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
	cwr             *countingWriter
	BaseFilename    string
	currentFilename string
	mutex           *sync.RWMutex
	RotateChan      chan string // defaults to nil.  receives previous filename on rotate
	RotateSize      int64       // rotate after RotateSize bytes have been written to the file

	rotateTicker *time.Ticker
	rotateReset  chan int
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
	// No lock here
	name := preprocessFilename(w.BaseFilename)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("TIMBER! Can't open %v: %v", name, err)
	}

	var cwr = &countingWriter{file, 0}
	var output io.WriteCloser = cwr
	// Wrap in gz writer
	if strings.HasSuffix(name, ".gz") {
		output = &gzFileWriter{
			gzip.NewWriter(output),
			output,
		}
	}

	// Locked from here
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.wr != nil {
		w.wr.Close()
		// send previous filename on rotate chan
		if c := w.RotateChan; c != nil {
			c <- w.currentFilename
		}
	}
	w.currentFilename = name
	w.cwr = cwr
	w.wr, _ = NewBufferedWriter(output)

	return nil
}

func (w *FileWriter) LogWrite(m string) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	w.wr.LogWrite(m)
	w.checkSize()
}

func (w *FileWriter) Flush() error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	e := w.wr.Flush()
	w.checkSize()
	return e
}

// Close and re-open the file.
// You should use the timestamp in the filename if you're going to use rotation
func (w *FileWriter) Rotate() error {
	return w.open()
}

// Automatically rotate every `d`
func (w *FileWriter) RotateEvery(d time.Duration) {
	// reset ticker
	w.mutex.Lock()
	if w.rotateTicker != nil {
		w.rotateTicker.Stop()
		w.rotateReset <- 1
	}
	w.rotateTicker = time.NewTicker(d)
	w.mutex.Unlock()

	// trigger a rotate every X
	go func() {
		for {
			select {
			case <-w.rotateReset:
				return
			case <-w.rotateTicker.C:
				w.Rotate()
			}
		}
	}()
}

func (w *FileWriter) checkSize() {
	if w.RotateSize > 0 && w.cwr.bytesWritten >= w.RotateSize {
		go w.Rotate()
	}
}

func (w *FileWriter) Close() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.wr.Close()
	w.wr = nil
}

type countingWriter struct {
	io.WriteCloser
	bytesWritten int64
}

func (w *countingWriter) Write(b []byte) (int, error) {
	i, e := w.WriteCloser.Write(b)
	w.bytesWritten += int64(i)
	return i, e
}

type gzFileWriter struct {
	*gzip.Writer // the compressor
	file         io.WriteCloser
}

func (w *gzFileWriter) Close() error {
	w.Writer.Close()
	return w.file.Close()
}
