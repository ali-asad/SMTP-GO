package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func handleClient(conn net.Conn) {
	defer conn.Close()

	// Send 220 greeting message
	conn.Write([]byte("220 example.com ESMTP Service ready\r\n"))

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		// Read the client's command
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading from client:", err)
			return
		}

		command = strings.TrimSpace(command)
		fmt.Println("Client command:", command)

		// Handle the command
		switch {
		case strings.HasPrefix(command, "HELO"):
			writer.WriteString("250 Hello\r\n")
		case strings.HasPrefix(command, "EHLO"):
			writer.WriteString("250-Hello\r\n250-SIZE 1000000\r\n")
		case strings.HasPrefix(command, "MAIL FROM"):
			writer.WriteString("250 OK\r\n")
		case strings.HasPrefix(command, "RCPT TO"):
			writer.WriteString("250 OK\r\n")
		case strings.HasPrefix(command, "DATA"):
			writer.WriteString("354 Start mail input; end with <CRLF>.<CRLF>\r\n")
			// Read the email data until a line containing only "."
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Error reading email data:", err)
					return
				}
				writer.WriteString(line)
				if line == ".\r\n" {
					break
				}
			}
			writer.WriteString("250 OK\r\n")
		case strings.HasPrefix(command, "QUIT"):
			writer.WriteString("221 Bye\r\n")
			writer.Flush()
			return
		default:
			writer.WriteString("500 Command not recognized\r\n")
		}

		writer.Flush()
	}
}

func main() {
	listener, err := net.Listen("tcp", ":25")
	if err != nil {
		log.Fatal("Error starting SMTP server:", err)
	}
	defer listener.Close()

	log.Println("SMTP server listening on port 25")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleClient(conn)
	}
}
