package ws

import (
	"strings"
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

type Conn struct {
	Ws   *websocket.Conn
	Send chan []byte
	ID   string
}

func newConn(ws *websocket.Conn) *Conn {
	c := &Conn{
		Send: make(chan []byte, 256),
		Ws:   ws,
		ID:   uuid.NewV4().String(),
	}
	c.Send <- []byte("hello|" + c.ID)
	return c
}

func (c *Conn) handleMessage(msg string) {
	if strings.HasPrefix(msg, "m|") {
	}
	return
}

func (c *Conn) write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

func (c *Conn) readPump() {
	defer func() {
		hub.unregisterConn <- c
		c.Ws.Close()
	}()
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msgType, rawMsg, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logrus.Warn("Unexpected ws close:", err)
			}
			break
		}

		if msgType == websocket.TextMessage {
			c.handleMessage(string(rawMsg))
		}
	}
}

func (c *Conn) writePump() {
	pingTimer := time.NewTicker(pingInterval)
	heartbeatTimer := time.NewTicker(heartbeatInterval)

	defer func() {
		pingTimer.Stop()
		heartbeatTimer.Stop()
		c.Ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			{
				if !ok {
					c.write(websocket.CloseMessage, []byte{})
					return
				}

				c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
				w, err := c.Ws.NextWriter(websocket.TextMessage)
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
				c.Send <- []byte("h")
			}
		}
	}
}
