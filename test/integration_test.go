package loadbalancer

import (
	"net/http"
	"sync"
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
	be := backend.Backend{Hostname: "127.0.0.1", Port: "8081"}
	go be.Start()
	defer be.Shutdown()

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

type RequestsCounter struct {
	count int
}

func (c *RequestsCounter) Increment() {
	c.count++
}

func TestMultipleBackends(t *testing.T) {
	counter1 := &RequestsCounter{}
	be := backend.Backend{Hostname: "127.0.0.1", Port: "8081", Counter: counter1}
	go be.Start()
	defer be.Shutdown()

	counter2 := &RequestsCounter{}
	be2 := backend.Backend{Hostname: "127.0.0.1", Port: "8082", Counter: counter2}
	go be2.Start()
	defer be2.Shutdown()

	counter3 := &RequestsCounter{}
	be3 := backend.Backend{Hostname: "127.0.0.1", Port: "8083", Counter: counter3}
	go be3.Start()
	defer be3.Shutdown()

	lb := loadbalancer.LoadBalancer{}
	go lb.Start()
	defer lb.Shutdown()

	// Wait for it to start
	time.Sleep(1 * time.Second)

	assertRequestIsSuccessful(t)
	assertRequestIsSuccessful(t)
	assertRequestIsSuccessful(t)

	expected := 1

	if counter1.count != 1 {
		t.Errorf("expected counter1 to be %v but got %v", expected, counter1.count)
	}
	if counter2.count != 1 {
		t.Errorf("expected counter2 to be %v but got %v", expected, counter2.count)
	}
	if counter3.count != 1 {
		t.Errorf("expected counter3 to be %v but got %v", expected, counter3.count)
	}
}

func TestRedirectsToAvailableServer(t *testing.T) {
	counter1 := &RequestsCounter{}
	be := backend.Backend{Hostname: "127.0.0.1", Port: "8081", Counter: counter1}
	go be.Start()
	defer be.Shutdown()

	counter2 := &RequestsCounter{}
	be2 := backend.Backend{Hostname: "127.0.0.1", Port: "8082", Counter: counter2}
	go be2.Start()

	lb := loadbalancer.LoadBalancer{}
	go lb.Start()
	defer lb.Shutdown()

	// Wait for it to start
	time.Sleep(1 * time.Second)
	be2.Shutdown() // disconnects server 2

	assertRequestIsSuccessful(t)
	assertRequestIsSuccessful(t)

	if counter1.count != 2 {
		t.Errorf("expected counter1 to be %v but got %v", 2, counter1.count)
	}
	if counter2.count != 0 {
		t.Errorf("expected counter2 to be %v but got %v", 0, counter2.count)
	}
}

func TestHandlesMultipleClientsAtSameTime(t *testing.T) {
	counter1 := &RequestsCounter{}
	be := backend.Backend{Hostname: "127.0.0.1", Port: "8081", Counter: counter1}
	go be.Start()
	defer be.Shutdown()

	counter2 := &RequestsCounter{}
	be2 := backend.Backend{Hostname: "127.0.0.1", Port: "8082", Counter: counter2}
	go be2.Start()

	lb := loadbalancer.LoadBalancer{}
	go lb.Start()
	defer lb.Shutdown()

	// Wait for it to start
	time.Sleep(1 * time.Second)

	assertParallelRequestsAreSuccessful(t)

	got := counter1.count + counter2.count
	expected := 10
	if got != expected {
		t.Errorf("expected %v but got %v", expected, got)
	}
}

func assertParallelRequestsAreSuccessful(t testing.TB) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go wgAssertRequestIsSuccessful(t, &wg)
	}

	wg.Wait()
}

func wgAssertRequestIsSuccessful(t testing.TB, wg *sync.WaitGroup) {
	assertRequestIsSuccessful(t)
	wg.Done()
}

func assertRequestIsSuccessful(t testing.TB) {
	t.Helper()

	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("Failed: %s", err)
	}

	got := resp.StatusCode
	expected := http.StatusOK

	if got != expected {
		t.Errorf("expected %v but got %v", expected, got)
	}
}
