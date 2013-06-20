package timber

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// This should write to anything that you can write to with net.Dial
type SocketWriter struct {
	conn        net.Conn
	network     string
	addr        string
	connSync    *sync.RWMutex
	restartOnce *sync.Once
}

func NewSocketWriter(network, addr string) (*SocketWriter, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return &SocketWriter{conn, network, addr, &sync.RWMutex{}, &sync.Once{}}, nil
}

func (sw *SocketWriter) LogWrite(msg string) {
	sw.connSync.RLock()
	_, err := sw.conn.Write([]byte(msg))
	sw.connSync.RUnlock()
	if err != nil {
		fmt.Printf("Socket logging error: %v", err)
		sw.restartOnce.Do(func() {
			go sw.reconnect()
		})
	}
}

func (sw *SocketWriter) reconnect() {
	for {
		conn, err := net.Dial(sw.network, sw.addr)
		if err == nil {
			sw.connSync.Lock()
			sw.conn = conn
			sw.restartOnce = &sync.Once{}
			sw.connSync.Unlock()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (sw *SocketWriter) Close() {
	sw.conn.Close()
}
