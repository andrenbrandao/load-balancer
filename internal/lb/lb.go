package loadbalancer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*

How to keep a list of servers and remove them if they are unhealthy?

- Keep a list of the servers we have to connect to and a list
of the active ones
- Keep iterating over the configuredServers and checking if they are healthy,
if one becomes inactive, remove it from the active list
- If they become active again, add them to the list

What happens if we run this health check in parallel and remove items from the list?
The main thread may be trying to access the array with the active servers and based
on the position, it may end up accessing an invalid index, raising a index out of range error.

So, this idea works if we keep executing this in the main thread. But, as soon
as we start creating goroutines, we get incorrect memory accesses.

How can we solve it?
-- First option
Keep a list of structs representing the servers with an active flag.
Iterate over the list and if it is active, try to handle that connection.
The health checks should only change that flag.

Drawbacks: if we have a big list of servers and only the first and last are active,
the round robin algorithm will have to iterate over the whole list to check which
nodes are active

*/

/*

When trying to do a load test with wrk and wrk2 I ended up getting many errors.

Wrk would report multiple socket errors and the load balancer would also exit
because of an EOF error.

Found out that the WRK request was sending a \0 byte. So, in that case I had
to close the request if any errors were found at reading.

Also, had to answer with HTTP Header Connection: close so that the clients
would expect the connection to be closed. Otherwise, I believe they
wanted to keep it alive.

But, if we use the http's package ListenAndServe, wrk can execute 200k
requests per second. While this load balancer, if it only returns a fixed HTTP
response, can only answer at 50k request/sec. Why is that? Maybe
it is because of the keep alive?

*/

const timeout = 5 * time.Second

type LoadBalancer struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	servers  []*server
}

type server struct {
	address string
	active  bool
}

func (s *server) activate() {
	s.active = true
}

func (s *server) deactivate() {
	s.active = false
}

func (lb *LoadBalancer) Start() {
	fmt.Println("Starting load balancer...")
	lb.quit = make(chan interface{})
	lb.servers = []*server{
		{address: "127.0.0.1:8081", active: true},
		{address: "127.0.0.1:8082", active: true},
		{address: "127.0.0.1:8083", active: true},
	}

	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	lb.listener = ln

	fmt.Fprintln(os.Stdout, "Listening for connections on 127.0.0.1:8080...")

	// go lb.checkHealthyServers()
	lb.acceptRequests(ln)
}

// Gracefully shutdown
// https://stackoverflow.com/a/66755998
func (lb *LoadBalancer) Shutdown() {
	fmt.Println("Shutting down...")
	close(lb.quit)
	lb.listener.Close()
	lb.wg.Wait()
	fmt.Println("Closed")
}

func (lb *LoadBalancer) acceptRequests(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		fmt.Println("Accepting request...")
		if err != nil {
			select {
			case <-lb.quit:
				log.Println("Closing listener...")
				return
			default:
				log.Fatal(err)
			}
		}

		lb.wg.Add(1)
		go lb.handleConnection(conn)
		fmt.Println()
	}
}

var serverPos int = -1

func (lb *LoadBalancer) getNextServer() (*server, error) {
	inactiveCount := 0

	serverPos = (serverPos + 1) % len(lb.servers)
	for !lb.servers[serverPos].active {
		serverPos = (serverPos + 1) % len(lb.servers)
		inactiveCount++

		if inactiveCount == len(lb.servers) {
			return nil, errors.New("all servers are down")
		}
	}

	return lb.servers[serverPos], nil
}

func (lb *LoadBalancer) handleConnection(conn net.Conn) {
	fmt.Printf("Received request from %s\n", conn.RemoteAddr())
	defer lb.wg.Done()
	defer conn.Close()
	fmt.Println("Connecting to a backend server...")

	var beConn net.Conn
	var srv *server
	for {
		var err error
		srv, err = lb.getNextServer()
		if err != nil {
			log.Println(err)
			buf := bytes.Buffer{}
			buf.WriteString("HTTP/1.1 502 Bad Gateway\r\n")
			buf.WriteString("Connection: close\r\n")
			conn.Write(buf.Bytes())
			conn.Close()
			return
		}

		beConn, err = net.DialTimeout("tcp", srv.address, timeout)
		if err != nil {
			log.Println(err)
			srv.deactivate()
			continue
		}
		break
	}
	defer beConn.Close()

	// Reuses the same connection while EOF is not found.
	// TODO:
	// [X] Reuse same client connection
	// [ ] Reuse same backend connection per client
	// [ ] Look into setKeepAlive method for TCP in Go. What does it do?
	for {
		fmt.Println("Reading from client...")
		clientReq, err := readFromConnection(conn)
		if err != nil {
			log.Println(err)
			log.Println("Closing connection...")
			conn.Close()
			return
		}

		fmt.Printf("Request from client %s: \n--\n%s\n--\n", conn.RemoteAddr(), clientReq)

		for {
			_, err = beConn.Write([]byte(clientReq))
			if err != nil {
				log.Println(err)
				srv.deactivate()
				continue
			}

			fmt.Println("Reading from backend...")
			backendRes, err := readFromConnection(beConn)
			if err != nil {
				log.Println(err)

				srv.deactivate()

				buf := bytes.Buffer{}
				buf.WriteString("HTTP/1.1 502 Bad Gateway\r\n")
				buf.WriteString("Connection: close\r\n")
				conn.Write(buf.Bytes())
				conn.Close()
				return
			}

			fmt.Printf("Response from server %s: \n--\n%s\n--\n", srv.address, backendRes)
			conn.Write([]byte(backendRes))
			break
		}
	}
}

func readFromConnection(conn net.Conn) (string, error) {
	fmt.Println("Reading from connection... " + conn.RemoteAddr().String())
	conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReader(conn)

	buf := bytes.Buffer{}
	contentLength := 0
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		buf.WriteString(s)

		if strings.HasPrefix(s, "Content-Length:") {
			lengthInStr := strings.Split(s, ":")[1]
			contentLength, err = strconv.Atoi(strings.TrimSpace(lengthInStr))
			if err != nil {
				log.Fatal(err)
			}
		}

		if s == "\r\n" {
			break
		}
	}

	for contentLength != 0 {
		b, err := reader.ReadByte()
		if err != nil {
			log.Fatal(err)
		}

		buf.WriteByte(b)
		contentLength--
	}

	return buf.String(), nil
}

func (lb *LoadBalancer) checkHealthyServers() {
	for {
		time.Sleep(10 * time.Second)
		for _, server := range lb.servers {
			if isHealthy(server.address) {
				server.activate()
			} else {
				server.deactivate()
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func isHealthy(serverAddress string) bool {
	beConn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("Could not connected to server %s\n", serverAddress)
		return false
	}

	buf := bytes.Buffer{}
	buf.WriteString("GET /health HTTP/1.1\r\n")
	buf.WriteString("\r\n")
	beConn.Write(buf.Bytes())

	s := fmt.Sprintf("Response from Health Check in server %s: ", serverAddress)
	res, err := readFromConnection(beConn)
	if err != nil {
		return false
	}

	fmt.Fprint(os.Stdout, s+res)
	beConn.Close()

	tokens := strings.Split(res, " ")
	if tokens[1] != "200" && tokens[1] != "204" {
		return false
	}

	return true
}
