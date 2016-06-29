package ws

var hub *hubType

func init() {
	hub = newHub()
	go hub.run()
}

type hubType struct {
	connections map[*Conn]struct{}
	handlers    map[string]MessageHandlerFunc

	broadcast         chan []byte
	registerConn      chan *Conn
	unregisterConn    chan *Conn
	registerHandler   chan messageHandler
	unregisterHandler chan string

	stop chan struct{}
}

type MessageHandlerFunc func(data []byte, reply chan []byte)

type messageHandler struct {
	messageName string
	handlerFunc MessageHandlerFunc
}

func newHub() *hubType {
	return &hubType{
		connections: make(map[*Conn]struct{}),
		handlers:    make(map[string]MessageHandlerFunc),

		broadcast:         make(chan []byte),
		registerConn:      make(chan *Conn),
		unregisterConn:    make(chan *Conn),
		registerHandler:   make(chan messageHandler),
		unregisterHandler: make(chan string),

		stop: make(chan struct{}),
	}
}

func (h *hubType) run() {
	for {
		select {
		case c := <-h.registerConn:
			h.connections[c] = struct{}{}

		case c := <-h.unregisterConn:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.Send)
			}

		case message := <-h.broadcast:
			for c := range h.connections {
				c.Send <- message
			}

		case handler := <-h.registerHandler:
			h.handlers[handler.messageName] = handler.handlerFunc

		case msgName := <-h.unregisterHandler:
			delete(h.handlers, msgName)

		case <-h.stop:
			return
		}
	}
}

func Broadcast(message []byte) {
	hub.broadcast <- message
}

func RegisterHandler(msgName string, handler MessageHandlerFunc) {
	hub.registerHandler <- messageHandler{msgName, handler}
}

func UnregisterHandler(msgName string) {
	hub.unregisterHandler <- msgName
}
