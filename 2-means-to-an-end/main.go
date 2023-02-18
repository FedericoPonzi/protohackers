package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
)

// Keys returns the keys of the map m.
// The keys will be an indeterminate order.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

func main() {
	ln, err := net.Listen("tcp", ":1337")
	if err != nil {
		fmt.Println("listen: ", err.Error())
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("connection from ", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	db := make(map[int32]int32)

	for {
		var msgType byte
		var v1, v2 int32
		err := binary.Read(conn, binary.BigEndian, &msgType)

		err = binary.Read(conn, binary.BigEndian, &v1)
		if err != nil {
			log.Printf("Error reading timestamp: %s", err)
			return
		}
		err = binary.Read(conn, binary.BigEndian, &v2)
		if err != nil {
			log.Printf("Error reading value: %s", err)
			return
		}

		if msgType == 'I' {
			// An insert message lets the client insert a timestamped price.

			timestamp := v1
			price := v2

			_, ok := db[timestamp]
			if ok {
				// Behaviour is undefined if there are multiple prices with the same timestamp from the same client.
				break
			}
			db[timestamp] = price

		} else if msgType == 'Q' {
			// A query message lets the client query the average price over a given time period.
			mintime := v1
			maxtime := v2

			// the sum can exceed int32
			var summed int64 = 0
			var count int64 = 0

			for timestamp, price := range db {
				if timestamp >= mintime && timestamp <= maxtime {
					summed += int64(price)
					count += 1
				}
			}

			var mean int32
			if count > 0 && mintime <= maxtime {
				mean = int32(summed / count)
			}

			err := binary.Write(conn, binary.BigEndian, mean)

			if err != nil {
				log.Printf("Error writing response: %s", err)
			} else {
				log.Printf("Sent mean price: %d", mean)
			}
			log.Printf("mintime: %d, maxtime: %d, summed: %d, count: %d\n", mintime, maxtime, summed, count)

		} else {
			// undefined behavior
			break
		}
		if err != nil {
			fmt.Println("accept: ", err.Error())
		}
	}
}
