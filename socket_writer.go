package timber

import (
	"net"
	"fmt"
)

// This should write to anything that you can write to with net.Dial
type SocketWriter struct {
	conn net.Conn
}

func NewSocketWriter(network, addr string) *SocketWriter {
	sw := new(SocketWriter)
	conn, err := net.Dial(network, addr)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't connect to socket: %v %v", network, addr))
	}
	sw.conn = conn
	return sw
}

func (sw *SocketWriter) LogWrite(msg string) {
	sw.conn.Write([]byte(msg))
}

func (sw *SocketWriter) Close() {
	sw.conn.Close()
}
//************/
/*************
func NewSocketWriter(network, addr string) (LogWriter) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		panic(fmt.Sprintf("TIMBER! Can't connect to socket: %v %v", network, addr))
	}
	return NewBufferedWriter(conn)
}
**************/
