package main

import loadbalancer "github.com/andrenbrandao/load-balancer/internal"

func main() {
	lb := loadbalancer.LoadBalancer{}
	lb.Start()
}
