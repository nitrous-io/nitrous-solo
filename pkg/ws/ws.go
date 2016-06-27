package ws

import "net/http"

func RunHub() {
	go hub.run()
}

func Upgrade(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	conn := newConn(ws)
	hub.register <- conn

	go conn.writePump()
	conn.readPump()
	return nil
}
