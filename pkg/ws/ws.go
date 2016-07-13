package ws

import "net/http"

type M map[string]interface{}

type MessageHandlerFunc func(c *Context)

type Context struct {
	Name  string
	Data  []byte
	Reply func([]byte)
	Meta  M
}

func Upgrade(w http.ResponseWriter, r *http.Request, meta map[string]interface{}) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c := newConn(ws)
	c.meta = meta
	c.send <- []byte("reg|" + c.id)
	hub.registerConn(c)

	go c.writePump()
	c.readPump()
	return nil
}

func Send(connID, msgID, name string, data []byte) {
	hub.send <- message{connID: connID, msgID: msgID, name: name, data: data}
}

func Broadcast(name string, data []byte) {
	hub.broadcast <- message{name: name, data: data}
}

func RegisterHandler(msgName string, handler MessageHandlerFunc) {
	hub.registerHandler(msgName, handler)
}

func UnregisterHandler(msgName string) {
	hub.unregisterHandler(msgName)
}
