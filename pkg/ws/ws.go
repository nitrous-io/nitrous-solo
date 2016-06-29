package ws

import "net/http"

func Upgrade(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c := newConn(ws)
	c.send <- []byte("reg|" + c.id)
	hub.registerConn <- c

	go c.writePump()
	c.readPump()
	return nil
}
