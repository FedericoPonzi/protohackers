package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"unicode"
)

const (
	MAX_MESSAGE_LENGTH = 1000
	MAX_NAME_LENGTH    = 16
)

var clients []*Client

// syncronize write access
var write sync.Mutex

type Client struct {
	name string
	conn net.Conn
}

func get_users(clients []*Client, exclude string) string {
	var result strings.Builder
	for i, s := range clients {
		if s.name == exclude {
			continue // skip this element
		}
		result.WriteString(s.name)
		if i != len(clients)-1 {
			result.WriteString(", ")
		}
	}
	fmt.Printf("user list: %s\n", result.String())
	return result.String()
}

func main() {
	ln, err := net.Listen("tcp", ":1337")
	if err != nil {
		fmt.Println("listen: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("listening on port 1337")
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
func isTaken(username string) bool {
	for _, c := range clients {
		if c.name == username {
			return true
		}
	}
	return false
}
func removeClient(client *Client) {
	write.Lock()
	for i, c := range clients {
		if c == client {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	write.Unlock()
	sendAll(fmt.Sprintf("* %s has left the room", client.name))
}

func handle(conn net.Conn) {
	defer func() {
		fmt.Println("Closing...")
		_ = conn.Close()
	}()
	_, err := fmt.Fprintf(conn, "%s\n", "Welcome to budgetchat! What shall I call you?")
	fmt.Println("message sent...")
	if err != nil {
		fmt.Printf("error %s ", err.Error())
		return
	}

	scanner := bufio.NewScanner(conn)
	fmt.Println("scanning...")
	var username string
	for scanner.Scan() {

		username = scanner.Text()

		if len(username) < 1 || len(username) > MAX_NAME_LENGTH || !isValidUsername(username) {
			fmt.Fprintf(conn, "%s", "Illegal name. Disconnecting.")
			return
		}

		// Check if the name is already taken
		if isTaken(username) {
			fmt.Fprintf(conn, "The name '%s' is already taken. Please choose a different name.\n", username)
			continue
		}

		break
	}
	if len(username) < 1 {
		fmt.Printf("error scan: %s\n", scanner.Err())
		return
	}
	fmt.Printf("User: '%s'\n", username)

	fmt.Printf("%s: sending the list\n", username)
	_, err = fmt.Fprintf(conn, "* The room contains: %s\n", get_users(clients, username))
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return
	}
	client := &Client{username, conn}
	clients = append(clients, client)

	// announce the new joiner
	sendAllExcept(fmt.Sprintf("* %s has entered the room", username), client)

	fmt.Println("starting loop...")
	for scanner.Scan() {
		// Read until newline
		message := scanner.Text()
		fmt.Println(message)

		write.Lock()
		sendAllExcept(fmt.Sprintf("[%s] %s", client.name, message), client)
		write.Unlock()
		// Do something with the line
	}
	removeClient(client)
}

func sendAllExcept(message string, except *Client) {
	for _, client := range clients {
		if client != except {
			_, err := fmt.Fprintf(client.conn, "%s\n", message)
			if err != nil {
				removeClient(client)
			}
		}
	}
}
func sendAll(message string) {
	for _, client := range clients {
		_, err := fmt.Fprintf(client.conn, "%s\n", message)
		if err != nil {
			removeClient(client)
		}
	}
}

func isAlphanumeric(s string) bool {
	hasAlphaNum := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlphaNum = true
		} else {
			return false
		}
	}
	return hasAlphaNum
}

func isValidUsername(username string) bool {
	return len(username) > 0 && isAlphanumeric(username) && len(username) <= 16
}
