package ws

import "net/http"

func Upgrade(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	conn := newConn(ws)
	hub.registerConn <- conn

	go conn.writePump()
	conn.readPump()
	return nil
}
