package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

func main() {

	// Sender data.
	from := "asad.mslm@outlook.com"
	password := "NSBKS-VYG7N-55EHH-MDGJF-GLFLF"

	// Receiver email address.
	to := []string{
		"asadalirana62@gmail.com",
		"saad.hassan@mslm.io",
		// "uman.shahzad@mslm.io",
	}

	// smtp server configuration.
	smtpHost := "mslm.email"
	smtpPort := "25"

	// Message.
	message := []byte("This is a test email message to test crms poc")

	// Authentication.
	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // InsecureSkipVerify should be false in production
		ServerName:         smtpHost,
	}

	// Sending email with TLS encryption.
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", smtpHost, smtpPort))
	if err != nil {
		fmt.Println("Error connecting to SMTP server:", err)
		return
	}
	defer client.Close()

	if err := client.StartTLS(tlsConfig); err != nil {
		fmt.Println("Error starting TLS:", err)
		return
	}

	if err := client.Auth(auth); err != nil {
		fmt.Println("Authentication error:", err)
		return
	}

	if err := client.Mail(from); err != nil {
		fmt.Println("Error setting sender:", err)
		return
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			fmt.Println("Error setting recipient:", err)
			return
		}
	}

	data, err := client.Data()
	if err != nil {
		fmt.Println("Error opening data connection:", err)
		return
	}
	defer data.Close()

	if _, err := data.Write(message); err != nil {
		fmt.Println("Error writing message:", err)
		return
	}

	fmt.Println("Email Sent Successfully!")

}
