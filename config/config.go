package config

import "os"

var (
	Env = "development"
)

func init() {
	if e := os.Getenv("NITROUS_ENV"); e != "" {
		Env = e
	}
}
