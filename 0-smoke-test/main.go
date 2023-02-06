package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:1337")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}
func handleRequest(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Handling connection..")
	buf, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading request.", err.Error())
	}
	_, err = conn.Write(buf)
	if err != nil {
		fmt.Println("Error writing response.", err.Error())
	}

}
