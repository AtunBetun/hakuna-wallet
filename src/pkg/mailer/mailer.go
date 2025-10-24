package mailer

import (
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

type MailDialer interface {
	DialAndSend(...*gomail.Message) error
}

func NewAppleMailDialer(host string, port int, username string, password string) MailDialer {
	d := gomail.NewDialer(host, port, username, password)
	return d

}

// SendAppleWalletEmail sends an email with a .pkpass Apple Wallet ticket attached and includes an HTML fallback link.
func SendAppleWalletEmail(
	from string,
	to string,
	subject string,
	dialer MailDialer,
	pkpassPath string,
) error {
	m := gomail.NewMessage()

	// Set headers
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	htmlBody := `
		<html>
		<body style="font-family: Helvetica, Arial, sans-serif; color: #333; font-size: 16px;">
			<p>Hi there,</p>
			<p>Thank you for your purchase! Your event ticket is attached below. You can add it directly to your Apple Wallet.</p>
			
			<p>Enjoy the event!<br>- The Team</p>
		</body>
		</html>
	`

	// Set HTML and plain-text alternative
	m.SetBody("text/html", htmlBody)

	// Attach .pkpass file with correct MIME type
	m.Attach(pkpassPath, gomail.SetHeader(map[string][]string{
		"Content-Type":              {"application/vnd.apple.pkpass"},
		"Content-Disposition":       {fmt.Sprintf(`attachment; filename="%s"`, "ticket.pkpass")},
		"Content-Transfer-Encoding": {"base64"},
	}))

	// Send
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send Apple Wallet email: %w", err)
	}

	return nil
}
