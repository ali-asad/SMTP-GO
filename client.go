package main

import (
	"fmt"
	"net/smtp"
)

func main() {
	// Define SMTP server configuration
	smtpServer := "localhost"
	smtpPort := 25
	from := "test@testemail.crmspoc.mslm.email"
	to := "asad.mslm@outlook.com"
	subject := "Test Email"
	body := "This is a test email from my SMTP server."

	// Compose the email message
	message := fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"\r\n" +
		body

	// Connect to the SMTP server
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", smtpServer, smtpPort))
	if err != nil {
		fmt.Println("Error connecting to SMTP server:", err)
		return
	}
	defer client.Close()

	// Send the email
	if err := client.Mail(from); err != nil {
		fmt.Println("Error setting sender:", err)
		return
	}
	if err := client.Rcpt(to); err != nil {
		fmt.Println("Error setting recipient:", err)
		return
	}
	writer, err := client.Data()
	if err != nil {
		fmt.Println("Error opening data connection:", err)
		return
	}
	_, err = writer.Write([]byte(message))
	if err != nil {
		fmt.Println("Error writing message body:", err)
		return
	}
	err = writer.Close()
	if err != nil {
		fmt.Println("Error closing data connection:", err)
		return
	}
	fmt.Println("Email sent successfully")
}
