package main

import (
	"flag"

	backend "github.com/andrenbrandao/load-balancer/internal/be"
)

func main() {
	var hostname, port string
	flag.StringVar(&hostname, "h", "127.0.0.1", "hostname")
	flag.StringVar(&port, "p", "8081", "port")
	flag.Parse()

	be := backend.Backend{Hostname: hostname, Port: port}
	be.Start()
}
