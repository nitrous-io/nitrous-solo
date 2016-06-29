package ws

func ResetHub_() {
	hub.stop <- struct{}{}
	hub = newHub()
	go hub.run()
}

func GetHub_() *hubType {
	return hub
}

func RegisterConn_(c *Conn) {
	hub.registerConn <- c
}

func UnregisterConn_(c *Conn) {
	hub.unregisterConn <- c
}

func HubConnections_() map[*Conn]struct{} {
	return hub.connections
}

func HubHandlers_() map[string]MessageHandlerFunc {
	return hub.handlers
}
