package main

import (
	"fmt"
	"net/smtp"
)

func main() {
	// Sender and receiver email addresses.
	from := "sender@testemail.crmspoc.mslm.email" // Replace with your sender email address.
	to := []string{"receiver@example.com"}        // Replace with the recipient's email address.

	// SMTP server configuration.
	smtpHost := "smtp.testemail.crmspoc.mslm.email"
	smtpPort := "587"

	// Message.
	message := []byte("Subject: Test Email\r\n" +
		"\r\n" +
		"This is a test email message.")

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, to, message)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}
