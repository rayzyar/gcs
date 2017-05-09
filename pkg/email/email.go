package email

import (
	"log"
	"net/smtp"
)

func Send(destEmail, body string) {
	from := "ray.spam2017@gmail.com"
	pass := "justspamit"
	to := "ray.zezhou@gmail.com"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: GCS Click the link to Confirm your food delivery\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("email sent")
}
