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
)

var serverAddresses []string = []string{"127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083"}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stdout, "Listening for connections on 127.0.0.1:8000...")
	serverPos := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, serverAddresses[serverPos])
		serverPos = (serverPos + 1) % len(serverAddresses)
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

	fmt.Printf("Response from server %s: ", serverAddress)
	buf = readFromConnection(beConn)
	fmt.Fprint(os.Stdout, buf.String())
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
