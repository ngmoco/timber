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
}

func NewBufferedWriter(writer io.WriteCloser) *BufferedWriter {
	bw := new(BufferedWriter)
	bw.writer = writer
	bw.buf = bufio.NewWriter(writer)
	return bw
}

func (bw *BufferedWriter) LogWrite(msg string) {
	_, err := bw.buf.Write([]byte(msg))
	if err != nil {
		// uh-oh... what do i do if logging fails; punt!
		log.Printf("TIMBER! epic fail: %v", err)
	}
}

// Force flush the buffer
func (bw *BufferedWriter) Flush() {
	bw.buf.Flush()
}

func (bw *BufferedWriter) Close() {
	bw.buf.Flush()
	bw.writer.Close()
}
