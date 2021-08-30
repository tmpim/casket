package httpserver

import (
	"context"
	"encoding/binary"
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tmpim/casket/caskethttp/tun"
	"github.com/xtaci/smux"
)

type TunSession struct {
	connQueue  chan net.Conn
	errorQueue chan error
	ctx        context.Context
	config     *tun.Config
}

func (t *TunSession) notifyError(err error) {
	select {
	case t.errorQueue <- err:
	default:
	}
}

func (t *TunSession) RunWorker() {
	for {
		if t.ctx.Err() != nil {
			return
		}

		conn, _, err := websocket.DefaultDialer.Dial(t.config.Upstream, http.Header{
			"Authorization": []string{"Bearer " + t.config.Secret},
		})
		t.notifyError(err)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		newSession, err := smux.Client(NewTunConn(conn), &smux.Config{
			Version:           1,
			KeepAliveInterval: 5 * time.Second,
			KeepAliveTimeout:  15 * time.Second,
			MaxFrameSize:      32768,
			MaxReceiveBuffer:  4194304,
			MaxStreamBuffer:   65536,
		})

		func() {
			defer conn.Close()

			for {
				stream, err := newSession.AcceptStream()
				if err != nil {
					log.Println("tun: error accepting stream:", err)
					return
					// TODO: write a proper logger
				}

				var headerSize uint16
				err = binary.Read(stream, binary.LittleEndian, &headerSize)
				if err != nil {
					log.Println("tun: error reading header:", err)
					stream.Close()
					continue
				}

				var addr net.Addr
				err = gob.NewDecoder(io.LimitReader(stream, int64(headerSize))).Decode(&addr)
				if err != nil {
					log.Println("tun: error parsing header:", err)
					stream.Close()
					continue
				}

				t.connQueue <- WrappedStream{
					Conn:       stream,
					remoteAddr: addr,
				}
			}
		}()

		return
	}
}

type WrappedStream struct {
	net.Conn
	remoteAddr net.Addr
}

func (s WrappedStream) RemoteAddr() net.Addr {
	return s.remoteAddr
}
