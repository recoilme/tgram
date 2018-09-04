package models

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strings"
)

func IsSmtpSet(SMTPHost, SMTPPort, SMTPUser, SMTPPassword string) bool {
	if SMTPHost == "" || SMTPPort == "" || SMTPUser == "" || SMTPPassword == "" {
		return false
	}
	return true
}

// SendMail send text/plain email via smtp
func SendMail(SMTPHost, SMTPPort, SMTPUser, SMTPPassword, Domain, address, title, body string) {
	if !IsSmtpSet(SMTPHost, SMTPPort, SMTPUser, SMTPPassword) {
		return
	}
	smtpServer := SMTPHost
	auth := smtp.PlainAuth(
		"",
		SMTPUser,
		SMTPPassword,
		smtpServer,
	)

	from := mail.Address{Domain, SMTPUser}
	to := mail.Address{address, address}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = encodeRFC2047(title)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		smtpServer+SMTPPort,
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)
	if err != nil {
		log.Println(err)
	}
}

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	// strange title in ""?
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <@>")
}
