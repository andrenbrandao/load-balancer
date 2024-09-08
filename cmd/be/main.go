package main

import backend "github.com/andrenbrandao/load-balancer/internal/be"

func main() {
	be := backend.Backend{}
	be.Start()
}
