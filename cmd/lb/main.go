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
*/

var serverAddresses []string = []string{"127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083"}
var activeServers []string = serverAddresses

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
		go checkHealthyServers()
		go handleConnection(conn, activeServers[serverPos])
		serverPos = (serverPos + 1) % len(activeServers)
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

func checkHealthyServers() {
	for _, serverAddress := range serverAddresses {
		if isHealthy(serverAddress) {
			activeServers = appendIfMissing(activeServers, serverAddress)
		} else {
			activeServers = removeIfPresent(activeServers, serverAddress)
		}
	}
	fmt.Print(activeServers)
}

func appendIfMissing(slice []string, s string) []string {
	for _, val := range slice {
		if val == s {
			return slice
		}
	}

	return append(slice, s)
}

func removeIfPresent(slice []string, s string) []string {
	ret := []string{}
	for _, val := range slice {
		if val != s {
			ret = append(ret, val)
		}
	}

	return ret
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

	fmt.Printf("Response from Health Check in server %s: ", serverAddress)
	buf = readFromConnection(beConn)
	fmt.Fprint(os.Stdout, buf.String())
	beConn.Close()

	tokens := strings.Split(buf.String(), " ")
	if tokens[1] != "200" && tokens[1] != "204" {
		return false
	}

	return true
}
