package ws

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

const (
	writeWait         = 120 * time.Second
	pongWait          = 60 * time.Second
	pingInterval      = pongWait * 9 / 10
	heartbeatInterval = 25 * time.Second
)

var upgrader = websocket.Upgrader{}

type Conn struct {
	ws   *websocket.Conn
	send chan []byte
}

func newConn(ws *websocket.Conn) *Conn {
	return &Conn{
		send: make(chan []byte, 256),
		ws:   ws,
	}
}

func (c *Conn) readPump() {
	defer func() {
		hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logrus.Warn("Unexpected ws close:", err)
			}
			break
		}
		hub.broadcast <- message
	}
}

func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) writePump() {
	pingTimer := time.NewTicker(pingInterval)
	heartbeatTimer := time.NewTicker(heartbeatInterval)

	defer func() {
		pingTimer.Stop()
		heartbeatTimer.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
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
				w.Write(message)

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
