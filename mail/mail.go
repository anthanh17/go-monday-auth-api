package mail

import (
	"log"
	"net/smtp"
	"strconv"
)

func SendEmail(receiverMail string, otpCode string) {
	// Set up authentication information.
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	senderEmail := "nonreply-otp@aihoply.com"
	senderPassword := "mfbi mkjr ysbd aurj"

	// Recipient email address.
	to := []string{receiverMail}

	// Email content.
	subject := "OTP Login Assistant DAG"
	body := "Your otp code is: " + otpCode

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
