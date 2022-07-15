package web

// SendEmail can email the toEmail from the fromEmail
/*func SendEmail(toEmail string, fromEmail string, subject string, body string, cfg *config.Configs) error {
	m := gomail.NewMessage()

	// receivers
	m.SetHeader("To", toEmail)

	// sender
	m.SetHeader("From", fromEmail)

	// subject
	m.SetHeader("Subject", subject)

	// E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", body)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, cfg.Configs.Settings.Email, cfg.Configs.Settings.EmailPassword)

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}*/
