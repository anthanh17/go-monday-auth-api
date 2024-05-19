package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
)

func SendEmail(userName string, receiverMail string, otpCode string) {
	// Set up authentication information.
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	senderEmail := "nonreply-otp@aihoply.com"
	senderPassword := "mfbi mkjr ysbd aurj"

	// Recipient email address.
	to := []string{receiverMail}

	// Email content.
	subject := "OTP Login Assistant DAG"

	body := fmt.Sprintf("Dear %s,\n\nYour OTP code is: %s\n\nPlease use this code to complete your verification. It is valid for 60s.\n\nIf you did not request this, please ignore this email.\n\nThank you,\nAihoply", userName, otpCode)

	// Message.
	message := "Subject: " + subject + "\r\n" +
		"\r\n" + body

	// Authentication.
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+strconv.Itoa(smtpPort), auth, senderEmail, to, []byte(message))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Email sent successfully!")
}
