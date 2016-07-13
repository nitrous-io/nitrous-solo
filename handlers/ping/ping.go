package pinghandler

import "github.com/nitrous-io/nitrous-solo/pkg/ws"

func Ping(c *ws.Context) {
	if len(c.Data) == 0 {
		c.Reply([]byte("pong"))
	} else {
		c.Reply(c.Data)
	}
}
