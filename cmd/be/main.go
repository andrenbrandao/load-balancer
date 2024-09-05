package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const timeout = 5 * time.Second

func main() {
	var hostname, port string
	flag.StringVar(&hostname, "h", "127.0.0.1", "hostname")
	flag.StringVar(&port, "p", "8081", "port")
	flag.Parse()

	ln, err := net.Listen("tcp", hostname+":"+port)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "Listening for connections on %s:%s...\n", hostname, port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
		fmt.Println()
	}
}

func handleConnection(conn net.Conn) {
	fmt.Fprintf(os.Stdout, "Received request from %s\n", conn.RemoteAddr())

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
