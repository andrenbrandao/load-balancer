package backend

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Counter interface {
	Increment()
}

const timeout = 5 * time.Second

type NullCounter struct{}

func (c *NullCounter) Increment() {
}

type Backend struct {
	Hostname string
	Port     string
	Counter  Counter // counts the number of requests received. Needed for testing.
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
}

func (b *Backend) Start() {
	fmt.Printf("Starting backend... Hostname: %s, Port: %s\n", b.Hostname, b.Port)
	ln, err := net.Listen("tcp", b.Hostname+":"+b.Port)
	if err != nil {
		log.Fatal(err)
	}

	b.quit = make(chan interface{})
	b.listener = ln

	if b.Counter == nil {
		b.Counter = &NullCounter{}
	}

	fmt.Fprintf(os.Stdout, "Listening for connections on %s:%s...\n", b.Hostname, b.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-b.quit:
				log.Println("Closing listener...")
				return
			default:
				log.Fatal(err)
			}
		}
		b.wg.Add(1)
		go b.handleConnection(conn)
		fmt.Println()
	}
}

func (b *Backend) Shutdown() {
	fmt.Println("Shutting down...")
	close(b.quit)
	b.listener.Close()
	b.wg.Wait()
	fmt.Println("Shut!")
}

func (b *Backend) handleConnection(conn net.Conn) {
	fmt.Fprintf(os.Stdout, "Received request from %s\n", conn.RemoteAddr())
	b.Counter.Increment()
	defer b.wg.Done()

	defer conn.Close()
	for {
		path, err := readPathFromConnection(conn)
		if err != nil {
			log.Println(err)
			return
		}
		buf := handleRoute(path)

		fmt.Printf("Response to client %s: \n--\n%s\n--\n", conn.RemoteAddr(), buf.Bytes())
		conn.Write(buf.Bytes())
	}
}

func readPathFromConnection(conn net.Conn) (string, error) {
	fmt.Println("Reading path from connection... " + conn.RemoteAddr().String())
	conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReader(conn)

	var path string
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		if s == "\r\n" {
			break
		}

		tokens := strings.Split(s, " ")
		if tokens[0] == "GET" {
			path = tokens[1]
		}

		fmt.Print(s)
	}
	fmt.Println()

	return path, nil
}

func handleRoute(path string) *bytes.Buffer {
	buf := bytes.Buffer{}

	switch path {
	case "/health":
		buf.Write([]byte("HTTP/1.1 204 No Content\r\n"))
		// buf.Write([]byte("Connection: close\r\n"))
		buf.Write([]byte("\r\n"))
	case "/":
		buf.Write([]byte("HTTP/1.1 200 OK\r\n"))
		// buf.Write([]byte("Connection: close\r\n"))
		buf.Write([]byte("Content-Length: 27\r\n"))
		buf.Write([]byte("\r\n"))
		buf.Write([]byte("Hello From Backend Server\r\n"))
	default:
		buf.Write([]byte("HTTP/1.1 404 Not Found\r\n"))
		// buf.Write([]byte("Connection: close\r\n"))
		buf.Write([]byte("\r\n"))
	}

	return &buf
}
