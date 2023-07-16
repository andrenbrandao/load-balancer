package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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

-- Second option

*/

type server struct {
	address string
	active  bool
}

var servers []*server = []*server{
	{address: "127.0.0.1:8081", active: true},
	{address: "127.0.0.1:8082", active: true},
	{address: "127.0.0.1:8083", active: true},
}

func (s *server) activate() {
	s.active = true
}

func (s *server) deactivate() {
	s.active = false
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stdout, "Listening for connections on 127.0.0.1:8000...")
	serverPos := 0

	go checkHealthyServers()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn, servers[serverPos].address)
		serverPos = (serverPos + 1) % len(servers)

		inactiveCount := 0
		for !servers[serverPos].active {
			serverPos = (serverPos + 1) % len(servers)
			inactiveCount++

			if inactiveCount == len(servers) {
				buf := bytes.Buffer{}
				buf.WriteString("HTTP/1.1 503 Service Unavailable\r\n")
				buf.WriteString("\r\n")
				conn.Write(buf.Bytes())
				conn.Close()
			}
		}
		fmt.Println()
	}

}

func handleConnection(conn net.Conn, serverAddress string) {
	fmt.Fprintf(os.Stdout, "Received request from %s\n", conn.RemoteAddr())
	buf := readFromConnection(conn)
	fmt.Fprint(os.Stdout, buf.String())

	beConn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatal(err)
	}
	beConn.Write(buf.Bytes())

	s := fmt.Sprintf("Response from server %s: ", serverAddress)
	buf = readFromConnection(beConn)
	fmt.Fprint(os.Stdout, s+buf.String())
	beConn.Close()

	conn.Write(buf.Bytes())
	conn.Close()
}

func readFromConnection(conn net.Conn) bytes.Buffer {
	reader := bufio.NewReader(conn)

	buf := bytes.Buffer{}
	contentLength := 0
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
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

	return buf
}

func checkHealthyServers() {
	for {
		for _, server := range servers {
			if isHealthy(server.address) {
				server.activate()
			} else {
				server.deactivate()
			}
			time.Sleep(200 * time.Millisecond)
		}
		time.Sleep(5 * time.Second)
	}
}

func isHealthy(serverAddress string) bool {
	beConn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("Could not connected to server %s", serverAddress)
		return false
	}

	buf := bytes.Buffer{}
	buf.WriteString("GET /health HTTP/1.1\r\n")
	buf.WriteString("\r\n")
	beConn.Write(buf.Bytes())

	s := fmt.Sprintf("Response from Health Check in server %s: ", serverAddress)
	buf = readFromConnection(beConn)
	fmt.Fprint(os.Stdout, s+buf.String())
	beConn.Close()

	tokens := strings.Split(buf.String(), " ")
	if tokens[1] != "200" && tokens[1] != "204" {
		return false
	}

	return true
}
