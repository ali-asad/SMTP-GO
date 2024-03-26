package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func client() {
	serverAddr := "localhost:25"

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatal("Error connecting to SMTP server:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read the server's greeting
	greeting, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading server greeting:", err)
	}
	fmt.Println("Server greeting:", greeting)

	// Send EHLO command
	writer.WriteString("EHLO example.com\r\n")
	writer.Flush()

	// Read response to EHLO
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading EHLO response:", err)
		}
		fmt.Println("EHLO response:", response)
		if response[:3] == "250" {
			break
		}
	}

	// Send MAIL FROM command
	writer.WriteString("MAIL FROM: <sender@example.com>\r\n")
	writer.Flush()

	// Read response to MAIL FROM
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading MAIL FROM response:", err)
	}
	fmt.Println("MAIL FROM response:", response)

	// Send RCPT TO command
	writer.WriteString("RCPT TO: <recipient@example.com>\r\n")
	writer.Flush()

	// Read response to RCPT TO
	response, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading RCPT TO response:", err)
	}
	fmt.Println("RCPT TO response:", response)

	// Send DATA command
	writer.WriteString("DATA\r\n")
	writer.Flush()

	// Read response to DATA
	response, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading DATA response:", err)
	}
	fmt.Println("DATA response:", response)

	// Send email body
	emailBody := "Subject: Test email\r\n\r\nThis is a test email sent from the SMTP client."
	writer.WriteString(emailBody + "\r\n.\r\n")
	writer.Flush()

	// Read response to email body
	response, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading email body response:", err)
	}
	fmt.Println("Email body response:", response)

	// Send QUIT command
	writer.WriteString("QUIT\r\n")
	writer.Flush()

	// Read response to QUIT
	response, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading QUIT response:", err)
	}
	fmt.Println("QUIT response:", response)
}
