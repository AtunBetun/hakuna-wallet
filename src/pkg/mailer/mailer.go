package mailer

import (
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

type MailDialer interface {
	DialAndSend(...*gomail.Message) error
}

func NewGoMailDialer(host string, port int, username string, password string) MailDialer {
	d := gomail.NewDialer(host, port, username, password)
	d.SSL = true
	return d

}

// SendAppleWalletEmail sends an email with a .pkpass Apple Wallet ticket attached and includes an HTML fallback link.
func SendAppleWalletEmail(
	from string,
	to string,
	subject string,
	dialer MailDialer,
	pkpassPath string,
	hostedPassURL string,
) error {
	m := gomail.NewMessage()

	// Set headers
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	// HTML body — uses Apple's badge asset as a fallback button.
	htmlBody := fmt.Sprintf(`
		<html>
		<body style="font-family: Helvetica, Arial, sans-serif; color: #333; font-size: 16px;">
			<p>Hi there,</p>
			<p>Thank you for your purchase! Your event ticket is attached below. You can add it directly to your Apple Wallet.</p>
			
			<!-- Fallback Add to Wallet button -->
			<p>If you don’t see the pass preview below, click the button:</p>
			<p>
				<a href="%s" target="_blank">
					<img src="https://developer.apple.com/wallet/images/add-to-apple-wallet-button.png"
						alt="Add to Apple Wallet"
						style="height: 45px; border: none;">
				</a>
			</p>

			<p>Enjoy the event!<br>- The Team</p>
		</body>
		</html>
	`, hostedPassURL)

	// Set HTML and plain-text alternative
	m.SetBody("text/html", htmlBody)
	m.AddAlternative("text/plain",
		fmt.Sprintf("Your ticket is attached. If you don’t see it, click here to add to Apple Wallet: %s", hostedPassURL),
	)

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
