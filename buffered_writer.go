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

	closeChan  chan bool
	closedChan chan bool
}

func NewBufferedWriter(writer io.WriteCloser) (*BufferedWriter, error) {
	bw := new(BufferedWriter)
	bw.writer = writer
	bw.buf = bufio.NewWriter(writer)
	bw.mc = make(chan string)
	bw.fc = make(chan int)
	bw.autoFlush = time.NewTicker(time.Second)
	bw.closeChan = make(chan bool)
	bw.closedChan = make(chan bool)
	go bw.writeLoop()
	return bw, nil
}

func (bw *BufferedWriter) writeLoop() {
	for {
		select {
		case msg := <-bw.mc:
			bw.writeMessage(msg)
		case <-bw.fc:
			bw.flush()
		case <-bw.autoFlush.C:
			bw.flush()
		case <-bw.closeChan:
			// close requested.  drain message queue and exit
			for {
				select {
				case msg := <-bw.mc:
					bw.writeMessage(msg)
				default:
					bw.flush()
					bw.writer.Close()
					close(bw.closedChan)
					return
				}
			}
		}
	}
}

func (bw *BufferedWriter) writeMessage(msg string) {
	_, err := bw.buf.Write([]byte(msg))
	if err != nil {
		// uh-oh... what do i do if logging fails; punt!
		fmt.Printf("TIMBER! epic fail: %v", err)
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
	select {
	case <-bw.closedChan:
		// writer is closed.  messages are discarded
	case bw.mc <- msg:
	}
}

// Force flush the buffer
func (bw *BufferedWriter) Flush() error {
	bw.fc <- 1
	return nil
}

func (bw *BufferedWriter) Close() {
	for {
		select {
		case bw.closeChan <- true:
		case <-bw.closedChan:
			return
		}
	}
}
