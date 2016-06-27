package ws

type Hub struct {
	connections map[*Conn]struct{}
	broadcast   chan []byte
	register    chan *Conn
	unregister  chan *Conn
}

var hub = Hub{
	broadcast:   make(chan []byte),
	register:    make(chan *Conn),
	unregister:  make(chan *Conn),
	connections: make(map[*Conn]struct{}),
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.connections[conn] = struct{}{}

		case conn := <-h.unregister:
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				close(conn.send)
			}

		case message := <-h.broadcast:
			for conn := range h.connections {
				select {
				case conn.send <- message:

				default:
					close(conn.send)
					delete(hub.connections, conn)
				}
			}
		}
	}
}
