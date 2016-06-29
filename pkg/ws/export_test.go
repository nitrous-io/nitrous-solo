package ws

func ResetHub() {
	hub.Stop <- struct{}{}
	hub = newHub()
	go hub.run()
}

func GetHub() *hubType {
	return hub
}
