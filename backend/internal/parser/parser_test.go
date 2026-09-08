package parser

import (
	"os"
	"testing"
)

func TestParseEmailWithLFBoundariesSubmittedOverSMTP(t *testing.T) {
	raw, err := os.ReadFile("../../email.eml")
	if err != nil {
		t.Fatal(err)
	}

	raw = append(raw, '\r', '\n')

	parsed, err := ParseEmail(raw)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Subject != "Shaks Test" {
		t.Fatalf("subject = %q", parsed.Subject)
	}
	if parsed.Text == "" || parsed.HTML == "" {
		t.Fatalf("expected both text and HTML bodies, got text=%q html=%q", parsed.Text, parsed.HTML)
	}
}
