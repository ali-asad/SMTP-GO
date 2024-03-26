package main

import (
	"fmt"
	"net/smtp"
)

func main() {

	// Sender data.
	from := "asad.mslm@"
	password := "sqav gskg iflx hyua"

	// Receiver email address.
	to := []string{
		"asadalirana62@gmail.com",
		"saad.hassan@mslm.io",
		"uman.shahzad@mslm.io",
	}

	// smtp server configuration.
	smtpHost := "testemail.crmspoc.mslm.email"
	smtpPort := "587"

	// Message.
	message := []byte("This is a test email message to test crms poc")

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}
