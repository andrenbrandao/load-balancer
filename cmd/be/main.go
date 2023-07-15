package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	fmt.Fprintf(os.Stdout, "Received request from %s\n", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if s == "\r\n" {
			break
		}
		fmt.Fprint(os.Stdout, s)
	}

	conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	conn.Write([]byte("Content-Length: 27\r\n"))
	conn.Write([]byte("\r\n"))
	conn.Write([]byte("Hello From Backend Server\r\n"))

	conn.Close()
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stdout, "Listening for connections...")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
		fmt.Println()
	}

}
