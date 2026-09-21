package test

import (
	"os"
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

func TestParseAttachmentsFromEML(t *testing.T) {
	raw, err := os.ReadFile("/home/grimreaper/projects/tempmail/backend/email.eml")
	if err != nil {
		t.Fatalf("read email.eml: %v", err)
	}

	atts, err := parser.ParseAttachments(raw)
	if err != nil {
		t.Fatalf("ParseAttachments() error = %v", err)
	}

	if len(atts) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(atts))
	}

	// Index 0 must be the text/plain attachment (was previously dropped).
	txt := atts[0]
	if txt.Filename != "test.txt" {
		t.Fatalf("att[0].Filename = %q, want %q", txt.Filename, "test.txt")
	}
	if txt.ContentType != "text/plain" {
		t.Fatalf("att[0].ContentType = %q, want %q", txt.ContentType, "text/plain")
	}
	if txt.Index != 0 {
		t.Fatalf("att[0].Index = %d, want 0", txt.Index)
	}
	if txt.Size == 0 {
		t.Fatal("att[0] has empty data")
	}
	if txt.Inline {
		t.Fatal("att[0] should not be inline")
	}

	// Index 1 must be the PDF and must be reachable.
	pdf := atts[1]
	if pdf.Filename != "test.pdf" {
		t.Fatalf("att[1].Filename = %q, want %q", pdf.Filename, "test.pdf")
	}
	if pdf.Index != 1 {
		t.Fatalf("att[1].Index = %d, want 1", pdf.Index)
	}
	if pdf.Size == 0 {
		t.Fatal("att[1] has empty data")
	}
	if string(pdf.Data[:4]) != "%PDF" {
		t.Fatalf("att[1] data does not start with %%PDF, got %q", string(pdf.Data[:min(len(pdf.Data), 16)]))
	}
	if pdf.Inline {
		t.Fatal("att[1] should not be inline")
	}
}

func TestParseInlineAttachment(t *testing.T) {
	raw := []byte("From: sender@example.com\r\n" +
		"Subject: Inline image\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/related; boundary=rel123\r\n" +
		"\r\n" +
		"--rel123\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		"<p>See <img src=\"cid:img1@example.com\" /></p>\r\n" +
		"--rel123\r\n" +
		"Content-Type: image/png\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"Content-ID: <img1@example.com>\r\n" +
		"Content-Disposition: inline\r\n" +
		"\r\n" +
		"iVBORw0KGgo=\r\n" +
		"--rel123--\r\n")

	atts, err := parser.ParseAttachments(raw)
	if err != nil {
		t.Fatalf("ParseAttachments() error = %v", err)
	}

	if len(atts) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(atts))
	}
	img := atts[0]
	if img.ContentID != "img1@example.com" {
		t.Fatalf("ContentID = %q, want %q", img.ContentID, "img1@example.com")
	}
	if !img.Inline {
		t.Fatal("expected inline=true for Content-Disposition: inline with content-id")
	}
	if img.ContentType != "image/png" {
		t.Fatalf("ContentType = %q, want image/png", img.ContentType)
	}
	if img.Size == 0 {
		t.Fatal("inline attachment has empty data")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
