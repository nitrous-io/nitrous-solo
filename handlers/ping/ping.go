package pinghandler

func Ping(name string, data []byte, reply func([]byte)) {
	if len(data) == 0 {
		reply([]byte("pong"))
	} else {
		reply(data)
	}
}
