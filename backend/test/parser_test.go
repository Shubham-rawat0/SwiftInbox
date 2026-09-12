package test

import (
	"testing"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/parser"
)

func TestParseEmailMultipartMessage(t *testing.T) {
	raw := []byte("From: Alice <alice@example.com>\r\n" +
		"Subject: Test message\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=boundary42\r\n" +
		"\r\n" +
		"--boundary42\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"Hello from the plain text part.\r\n" +
		"--boundary42\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		"<p>Hello from the <strong>HTML</strong> part.</p>\r\n" +
		"--boundary42--\r\n")

	parsed, err := parser.ParseEmail(raw)
	if err != nil {
		t.Fatalf("ParseEmail() error = %v", err)
	}

	if parsed.From != "Alice <alice@example.com>" {
		t.Fatalf("From = %q, want %q", parsed.From, "Alice <alice@example.com>")
	}
	if parsed.Subject != "Test message" {
		t.Fatalf("Subject = %q, want %q", parsed.Subject, "Test message")
	}
	if parsed.Text != "Hello from the plain text part." {
		t.Fatalf("Text = %q", parsed.Text)
	}
	if parsed.HTML != "<p>Hello from the <strong>HTML</strong> part.</p>" {
		t.Fatalf("HTML = %q", parsed.HTML)
	}
}
