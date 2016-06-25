package main

import "github.com/nitrous-io/nitrous-solo/server"

func main() {
	r := server.New()
	r.Run(":3000")
}
