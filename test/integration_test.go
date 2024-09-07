package loadbalancer

import (
	"net/http"
	"testing"
	"time"

	loadbalancer "github.com/andrenbrandao/load-balancer/internal"
)

// When there are no backend servers
func TestAnswers502BadGateway(t *testing.T) {
	lb := loadbalancer.LoadBalancer{}
	go lb.Start()

	// Wait for it to start
	time.Sleep(1 * time.Second)
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected %q but got %q", http.StatusBadGateway, resp.StatusCode)
	}

	lb.Shutdown()
}
