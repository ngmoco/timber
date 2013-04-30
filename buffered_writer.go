package timber

import (
	"bufio"
	"io"
	"log"
)

// Use this of you need some buffering, or not
type BufferedWriter struct {
	buf    *bufio.Writer
	writer io.WriteCloser
	mc     chan string
	fc     chan int
}

func NewBufferedWriter(writer io.WriteCloser) *BufferedWriter {
	bw := new(BufferedWriter)
	bw.writer = writer
	bw.buf = bufio.NewWriter(writer)
	bw.mc = make(chan string)
	bw.fc = make(chan int)
	go bw.writeLoop()
	return bw
}

func (bw *BufferedWriter) writeLoop() {
	for {
		select {
		case msg, ok := <-bw.mc:
			if !ok {
				bw.buf.Flush()
				bw.writer.Close()
				return
			}
			_, err := bw.buf.Write([]byte(msg))
			if err != nil {
				// uh-oh... what do i do if logging fails; punt!
				log.Printf("TIMBER! epic fail: %v", err)
			}
		case <-bw.fc:
			bw.buf.Flush()
		}
	}
}

func (bw *BufferedWriter) LogWrite(msg string) {
	bw.mc <- msg
}

// Force flush the buffer
func (bw *BufferedWriter) Flush() {
	bw.fc <- 1
}

func (bw *BufferedWriter) Close() {
	close(bw.mc)
}
