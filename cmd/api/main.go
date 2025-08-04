package main

import (
	"check24-techup25/api"
	"check24-techup25/logging"
	"flag"
)

func main() {
	logging.SetLogger()
	port := flag.Int("port", 8080, "Set the port of the API")
	api.Start(*port)
}
