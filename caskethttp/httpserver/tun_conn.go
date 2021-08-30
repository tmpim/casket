package httpserver

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type TunConn struct {
	buf  []byte
	conn *websocket.Conn
}

func NewTunConn(conn *websocket.Conn) *TunConn {
	return &TunConn{
		conn: conn,
	}
}

func (c *TunConn) Read(b []byte) (int, error) {
	if len(c.buf) > 0 {
		n := copy(b, c.buf)
		c.buf = c.buf[n:]
		return n, nil
	}

	for {
		t, data, err := c.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		if t != websocket.BinaryMessage {
			continue
		}

		c.buf = data

		return c.Read(b)
	}
}

func (c *TunConn) Write(b []byte) (int, error) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (c *TunConn) Close() error {
	return c.conn.Close()
}

func (c *TunConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TunConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TunConn) SetDeadline(t time.Time) error {
	// not supported
	return nil
}

func (c *TunConn) SetReadDeadline(t time.Time) error {
	// not supported
	return nil
}

func (c *TunConn) SetWriteDeadline(t time.Time) error {
	// not supported
	return nil
}
