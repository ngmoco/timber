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
	Timeout     time.Duration
}

func NewSocketWriter(network, addr string) (*SocketWriter, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	timeout := 5 * time.Millisecond // logging should be fast
	return &SocketWriter{conn, network, addr, &sync.RWMutex{}, &sync.Once{}, timeout}, nil
}

func (sw *SocketWriter) LogWrite(msg string) {
	sw.connSync.RLock()
	// Starting with go1.1 (currently tested on go1.1.1)
	// writing to /dev/log on linux with rsyslog will occasionally
	// return EAGAIN.  Unfortunately, the go socket code does
	// a blocking wait which never returns when it gets an EAGAIN,
	// so a timeout is necessary to unblock. I tested and a reconnect
	// is not necessary when this happens but I think the code is more
	// general this way so I'm leaving the reconnect on all errors.
	sw.conn.SetWriteDeadline(time.Now().Add(sw.Timeout))
	_, err := sw.conn.Write([]byte(msg))
	sw.connSync.RUnlock()
	if err != nil {
		fmt.Printf("Socket logging error: %v\n", err)
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
