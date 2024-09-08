package main

import loadbalancer "github.com/andrenbrandao/load-balancer/internal/lb"

func main() {
	lb := loadbalancer.LoadBalancer{}
	lb.Start()
}
