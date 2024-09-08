package loadbalancer

import (
	"net/http"
	"testing"
	"time"

	backend "github.com/andrenbrandao/load-balancer/internal/be"
	loadbalancer "github.com/andrenbrandao/load-balancer/internal/lb"
)

// When there are no backend servers
func TestAnswers502BadGateway(t *testing.T) {
	lb := loadbalancer.LoadBalancer{}
	go lb.Start()
	defer lb.Shutdown()

	// Wait for it to start
	time.Sleep(1 * time.Second)
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}

	got := resp.StatusCode
	expected := http.StatusBadGateway

	if got != expected {
		t.Errorf("expected %v but got %v", expected, got)
	}
}

func TestAnswers200WhenBackendIsRunning(t *testing.T) {
	be := backend.Backend{}
	go be.Start()
	lb := loadbalancer.LoadBalancer{}
	go lb.Start()
	defer lb.Shutdown()

	// Wait for it to start
	time.Sleep(1 * time.Second)
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}

	got := resp.StatusCode
	expected := http.StatusOK

	if got != expected {
		t.Errorf("expected %v but got %v", expected, got)
	}
}
