package main

import (
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	"github.com/nitrous-io/nitrous-solo/server"
)

func main() {
	go ws.RunHub()
	r := server.New()
	r.Run(":3000")
}
