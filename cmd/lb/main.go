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

	conn.Close()
}

func main() {

	ln, err := net.Listen("tcp", "127.0.0.1:8080")
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
