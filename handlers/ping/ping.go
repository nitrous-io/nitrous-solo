package pinghandler

import "github.com/nitrous-io/nitrous-solo/pkg/ws"

func Ping(c *ws.Context) {
	defer c.Done()

	if len(c.Data) == 0 {
		c.Reply([]byte("pong"))
	} else {
		c.Reply(c.Data)
	}
}
