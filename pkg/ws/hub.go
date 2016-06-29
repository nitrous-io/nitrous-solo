package ws

import "bytes"

var hub *hubType

func init() {
	hub = newHub()
	go hub.run()
}

type hubType struct {
	connections map[string]*conn
	handlers    map[string]MessageHandlerFunc

	send              chan message
	broadcast         chan message
	registerConn      chan *conn
	unregisterConn    chan *conn
	registerHandler   chan messageHandler
	unregisterHandler chan string
	handleMessage     chan message

	stop chan struct{}
}

func newHub() *hubType {
	return &hubType{
		connections: make(map[string]*conn),
		handlers:    make(map[string]MessageHandlerFunc),

		send:              make(chan message, 256),
		broadcast:         make(chan message, 256),
		registerConn:      make(chan *conn),
		unregisterConn:    make(chan *conn),
		registerHandler:   make(chan messageHandler),
		unregisterHandler: make(chan string),
		handleMessage:     make(chan message, 256),

		stop: make(chan struct{}),
	}
}

type message struct {
	connID string
	msgID  string
	name   string
	data   []byte
}

type MessageHandlerFunc func(name string, data []byte, reply func([]byte))

type messageHandler struct {
	msgName     string
	handlerFunc MessageHandlerFunc
}

func (h *hubType) run() {
	for {
		select {
		case c := <-h.registerConn:
			h.connections[c.id] = c

		case c := <-h.unregisterConn:
			if _, ok := h.connections[c.id]; ok {
				delete(h.connections, c.id)
				close(c.send)
			}

		case m := <-h.send:
			if c, ok := h.connections[m.connID]; ok {
				c.send <- append([]byte("m|"+m.name+"|"+m.msgID+"|"), m.data...)
			}

		case m := <-h.broadcast:
			for _, c := range h.connections {
				c.send <- append([]byte("m|"+m.name+"||"), m.data...)
			}

		case handler := <-h.registerHandler:
			h.handlers[handler.msgName] = handler.handlerFunc

		case msgName := <-h.unregisterHandler:
			delete(h.handlers, msgName)

		case m := <-h.handleMessage:
			if handler, ok := h.handlers[m.name]; ok {
				name := m.name
				connID := m.connID
				msgID := m.msgID

				go handler(name, m.data, func(replyData []byte) {
					Send(connID, msgID, name, replyData)
				})
			}

		case <-h.stop:
			return
		}
	}
}

func parseAndHandleMessage(conn *conn, m []byte) {
	if bytes.HasPrefix(m, []byte("m|")) {
		parts := bytes.SplitN(m, []byte("|"), 4)
		hub.handleMessage <- message{connID: conn.id, msgID: string(parts[2]), name: string(parts[1]), data: parts[3]}
	}
}

func Send(connID, msgID, name string, data []byte) {
	hub.send <- message{connID: connID, msgID: msgID, name: name, data: data}
}

func Broadcast(name string, data []byte) {
	hub.broadcast <- message{name: name, data: data}
}

func RegisterHandler(msgName string, handler MessageHandlerFunc) {
	hub.registerHandler <- messageHandler{msgName, handler}
}

func UnregisterHandler(msgName string) {
	hub.unregisterHandler <- msgName
}
