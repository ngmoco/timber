package timber

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

type flusher interface {
	Flush() error
}

// Use this of you need some buffering, or not
type BufferedWriter struct {
	buf       *bufio.Writer
	writer    io.WriteCloser
	mc        chan string
	fc        chan int
	autoFlush *time.Ticker
}

func NewBufferedWriter(writer io.WriteCloser) (*BufferedWriter, error) {
	bw := new(BufferedWriter)
	bw.writer = writer
	bw.buf = bufio.NewWriter(writer)
	bw.mc = make(chan string)
	bw.fc = make(chan int)
	bw.autoFlush = time.NewTicker(time.Second)
	go bw.writeLoop()
	return bw, nil
}

func (bw *BufferedWriter) writeLoop() {
	for {
		select {
		case msg, ok := <-bw.mc:
			if !ok {
				bw.flush()
				bw.writer.Close()
				return
			}
			_, err := bw.buf.Write([]byte(msg))
			if err != nil {
				// uh-oh... what do i do if logging fails; punt!
				fmt.Printf("TIMBER! epic fail: %v", err)
			}
		case <-bw.fc:
			bw.flush()
		case <-bw.autoFlush.C:
			bw.flush()
		}
	}
}

// perform actual flush.  only on writeLoop goroutine
func (bw *BufferedWriter) flush() {
	// flush buffer
	bw.buf.Flush()
	// flush underlying buffer if supported
	if f, ok := bw.writer.(flusher); ok {
		f.Flush()
	}
}

func (bw *BufferedWriter) LogWrite(msg string) {
	bw.mc <- msg
}

// Force flush the buffer
func (bw *BufferedWriter) Flush() error {
	bw.fc <- 1
	return nil
}

func (bw *BufferedWriter) Close() {
	close(bw.mc)
}
