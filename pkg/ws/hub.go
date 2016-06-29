package ws

var hub *hubType

func init() {
	hub = newHub()
	go hub.run()
}

type hubType struct {
	Connections    map[*Conn]struct{}
	Broadcast      chan []byte
	RegisterConn   chan *Conn
	UnregisterConn chan *Conn
	Stop           chan struct{}
}

func newHub() *hubType {
	return &hubType{
		Connections:    make(map[*Conn]struct{}),
		Broadcast:      make(chan []byte),
		RegisterConn:   make(chan *Conn),
		UnregisterConn: make(chan *Conn),
		Stop:           make(chan struct{}),
	}
}

func (h *hubType) run() {
	for {
		select {
		case c := <-h.RegisterConn:
			h.Connections[c] = struct{}{}

		case c := <-h.UnregisterConn:
			if _, ok := h.Connections[c]; ok {
				delete(h.Connections, c)
				close(c.Send)
			}

		case message := <-h.Broadcast:
			for c := range h.Connections {
				c.Send <- message
			}

		case <-h.Stop:
			return
		}
	}
}
