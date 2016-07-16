package ws

import (
	"bytes"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

const (
	writeWait         = 120 * time.Second
	pongWait          = 60 * time.Second
	pingInterval      = pongWait * 9 / 10
	heartbeatInterval = 25 * time.Second
)

var upgrader = websocket.Upgrader{}

type conn struct {
	ws           *websocket.Conn
	send         chan []byte
	id           string
	meta         map[string]interface{}
	onCloseChans map[chan struct{}]struct{}
}

func newConn(ws *websocket.Conn) *conn {
	return &conn{
		send:         make(chan []byte, 256),
		ws:           ws,
		id:           uuid.NewV4().String(),
		meta:         M{},
		onCloseChans: map[chan struct{}]struct{}{},
	}
}

func (c *conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *conn) readPump() {
	defer func() {
		hub.unregisterConn(c)
		c.ws.Close()
	}()
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msgType, rawMsg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logrus.Warn("Unexpected ws close:", err)
			}
			break
		}

		if msgType == websocket.TextMessage {
			parseAndHandleMessage(c, rawMsg)
		}
	}

	hub.lock.Lock()
	defer hub.lock.Unlock()
	for ch, _ := range c.onCloseChans {
		ch <- struct{}{}
		delete(c.onCloseChans, ch)
		close(ch)
	}
}

func parseAndHandleMessage(conn *conn, m []byte) {
	if bytes.HasPrefix(m, []byte("m|")) {
		parts := bytes.SplitN(m, []byte("|"), 4)
		hub.handleMessage <- message{connID: conn.id, msgID: string(parts[2]), name: string(parts[1]), data: parts[3]}
	}
}

func (c *conn) writePump() {
	pingTimer := time.NewTicker(pingInterval)
	heartbeatTimer := time.NewTicker(heartbeatInterval)

	defer func() {
		pingTimer.Stop()
		heartbeatTimer.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			{
				if !ok {
					c.write(websocket.CloseMessage, []byte{})
					return
				}

				c.ws.SetWriteDeadline(time.Now().Add(writeWait))
				w, err := c.ws.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(data)

				if err := w.Close(); err != nil {
					return
				}
			}

		case <-pingTimer.C:
			{
				if err := c.write(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}

		case <-heartbeatTimer.C:
			{
				c.send <- []byte("h")
			}
		}
	}
}
