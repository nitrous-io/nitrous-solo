package ws

import "sync"

var hub *hubType

func init() {
	hub = newHub()
	go hub.run()
}

type hubType struct {
	connections map[string]*conn
	handlers    map[string]MessageHandlerFunc
	lock        sync.Mutex

	send          chan message
	broadcast     chan message
	handleMessage chan message

	stop chan struct{}
}

func newHub() *hubType {
	return &hubType{
		connections: map[string]*conn{},
		handlers:    map[string]MessageHandlerFunc{},

		send:          make(chan message, 256),
		broadcast:     make(chan message, 256),
		handleMessage: make(chan message, 256),

		stop: make(chan struct{}),
	}
}

type message struct {
	connID string
	msgID  string
	name   string
	data   []byte
}

func (h *hubType) run() {
	for {
		select {
		case m := <-h.send:
			if c, ok := h.connections[m.connID]; ok {
				c.send <- append([]byte("m|"+m.name+"|"+m.msgID+"|"), m.data...)
			}

		case m := <-h.broadcast:
			for _, c := range h.connections {
				c.send <- append([]byte("m|"+m.name+"||"), m.data...)
			}

		case m := <-h.handleMessage:
			conn, ok := h.connections[m.connID]
			if !ok {
				break
			}

			if handler, ok := h.handlers[m.name]; ok {
				name := m.name
				data := m.data
				connID := m.connID
				msgID := m.msgID

				go handler(&Context{
					Name: name,
					Data: data,
					Reply: func(replyData []byte) {
						Send(connID, msgID, name, replyData)
					},
					Meta: conn.meta,
				})
			}

		case <-h.stop:
			close(h.send)
			close(h.broadcast)
			close(h.handleMessage)
			close(h.stop)
			return // exit function
		}
	}
}

func (h *hubType) registerConn(c *conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.connections[c.id] = c
}

func (h *hubType) unregisterConn(c *conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.connections[c.id]; ok {
		delete(h.connections, c.id)
		close(c.send)
	}
}

func (h *hubType) registerHandler(msgName string, handler MessageHandlerFunc) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.handlers[msgName] = handler
}

func (h *hubType) unregisterHandler(msgName string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.handlers, msgName)
}
