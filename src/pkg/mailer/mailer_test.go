package mailer

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gomail "gopkg.in/gomail.v2"
)

type mockDialer struct {
	messages []*gomail.Message
}

func (m *mockDialer) DialAndSend(msgs ...*gomail.Message) error {
	m.messages = append(m.messages, msgs...)
	return nil
}

func TestSendAppleWalletEmailBuildsMessage(t *testing.T) {
	const (
		from          = "from@example.com"
		to            = "to@example.com"
		subject       = "Your Ticket"
		hostedPassURL = "https://cdn.example.com/ticket.pkpass"
	)

	tmpDir := t.TempDir()
	pkpassPath := filepath.Join(tmpDir, "ticket.pkpass")
	if err := os.WriteFile(pkpassPath, []byte("pkpass content"), 0o600); err != nil {
		t.Fatalf("write pkpass: %v", err)
	}

	mock := &mockDialer{}

	if err := SendAppleWalletEmail(from, to, subject, mock, pkpassPath, hostedPassURL); err != nil {
		t.Fatalf("SendAppleWalletEmail returned error: %v", err)
	}

	if len(mock.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(mock.messages))
	}

	msg := mock.messages[0]

	assertHeader := func(key, want string) {
		values := msg.GetHeader(key)
		if len(values) != 1 || values[0] != want {
			t.Fatalf("header %s = %v, want %s", key, values, want)
		}
	}

	assertHeader("From", from)
	assertHeader("To", to)
	assertHeader("Subject", subject)

	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		t.Fatalf("write message: %v", err)
	}
	rendered := buf.String()

	expectedSnippets := []string{
		"Content-Type: text/html",
		hostedPassURL,
		"Add to Apple Wallet",
		"Content-Type: text/plain",
		"Content-Type: application/vnd.apple.pkpass",
		`Content-Disposition: attachment; filename="ticket.pkpass"`,
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered message to contain %q", snippet)
		}
	}
}

type failingDialer struct {
	err error
}

func (f failingDialer) DialAndSend(...*gomail.Message) error {
	return f.err
}

func TestSendAppleWalletEmailPropagatesDialerError(t *testing.T) {
	wantErr := "smtp error"
	tmpDir := t.TempDir()
	pkpassPath := filepath.Join(tmpDir, "ticket.pkpass")
	if err := os.WriteFile(pkpassPath, []byte("pkpass content"), 0o600); err != nil {
		t.Fatalf("write pkpass: %v", err)
	}

	err := SendAppleWalletEmail("from@example.com", "to@example.com", "subject", failingDialer{err: errors.New(wantErr)}, pkpassPath, "https://example.com")
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("expected error containing %q, got %v", wantErr, err)
	}
}
