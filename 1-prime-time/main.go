package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
)

type request struct {
	Method *string  `json:"method"`
	Number *float64 `json:"number"`
}

type response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:1337")
	if err != nil {
		fmt.Println("Error listening on port", err.Error())
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting", err.Error())
			break
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	//fmt.Println("handling request")
	// we create a decoder that reads directly from the socket
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing conn.", err.Error())
		}
	}(conn)

	bufReader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	respTrue, _ := json.Marshal(response{Method: "isPrime", Prime: true})
	respFalse, _ := json.Marshal(response{Method: "isPrime", Prime: false})

	for {
		readBytes, err := bufReader.ReadBytes('\n')
		if err != nil {
			fmt.Println("Error reading request.", err.Error())
			break
		}
		var reqMsg request
		err = json.Unmarshal(readBytes, &reqMsg)
		if err != nil || reqMsg.Method == nil || *reqMsg.Method != "isPrime" || reqMsg.Number == nil {
			//fmt.Println("Failed to decode")
			_, err := writer.Write([]byte("{malformed}"))
			if err != nil {
				fmt.Println("Failed to decode", err.Error())
			}
			break
		}
		//.Println("Received message: ", string(readBytes))
		if big.NewInt(int64(*reqMsg.Number)).ProbablyPrime(0) {
			fmt.Println("it's prime!")
			_, err := writer.Write(respTrue)
			if err != nil {
				fmt.Println("Error writing response.", err.Error())
				break
			}
		} else {
			fmt.Println("not prime: ", reqMsg.Number)
			_, err := writer.Write(respFalse)
			if err != nil {
				fmt.Println("Error writing response.", err.Error())
				break
			}
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			fmt.Println("Error writing response.", err.Error())
			break
		}
		writer.Flush()
	}

}
