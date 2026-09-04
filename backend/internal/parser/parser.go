package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
)

type Attachment struct {
	Filename    string
	ContentType string
	ContentID   string
	Size        int
	Data        []byte
}

func ParseTextBody(raw []byte) (string, error) {
	entity, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("read email: %w", err)
	}

	var plain, html string
	if err := walk(entity, func(contentType string, params map[string]string, data []byte) error {
		switch contentType {
		case "text/plain":
			if plain == "" {
				plain = string(data)
			}
		case "text/html":
			if html == "" {
				html = string(data)
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	if plain != "" {
		return plain, nil
	}
	if html != "" {
		return stripTags(html), nil
	}
	return "", nil
}

func ParseHTMLBody(raw []byte) (string, error) {
	entity, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("read email: %w", err)
	}

	var html string
	if err := walk(entity, func(contentType string, params map[string]string, data []byte) error {
		if contentType == "text/html" && html == "" {
			html = string(data)
		}
		return nil
	}); err != nil {
		return "", err
	}

	return html, nil
}

func ParseAttachments(raw []byte) ([]Attachment, error) {
	entity, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("read email: %w", err)
	}

	var attachments []Attachment
	if err := walk(entity, func(contentType string, params map[string]string, data []byte) error {
		if contentType == "text/plain" || contentType == "text/html" {
			return nil
		}

		filename := params["name"]
		contentID := strings.Trim(params["content-id"], "<>")

		attachments = append(attachments, Attachment{
			Filename:    filename,
			ContentType: contentType,
			ContentID:   contentID,
			Size:        len(data),
			Data:        data,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return attachments, nil
}

func walk(entity *message.Entity, fn func(contentType string, params map[string]string, data []byte) error) error {
	contentType, params, err := entity.Header.ContentType()
	if err != nil {
		return fmt.Errorf("parse content type: %w", err)
	}

	if strings.HasPrefix(contentType, "multipart/") {
		reader := entity.MultipartReader()
		if reader == nil {
			return fmt.Errorf("email is multipart but has no multipart reader")
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read MIME part: %w", err)
			}
			if err := walk(part, fn); err != nil {
				return err
			}
		}
		return nil
	}

	data, err := io.ReadAll(entity.Body)
	if err != nil {
		return fmt.Errorf("read MIME body: %w", err)
	}

	return fn(contentType, params, data)
}


func stripTags(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func MakePreview(text string, max int) string {
    text = strings.TrimSpace(text)

    runes := []rune(text)

    if len(runes) <= max {
        return text
    }

    return string(runes[:max]) + "..."
}


type ParsedEmail struct {
	From    string
	Subject string
	Text    string
	HTML    string
}

func ParseEmail(raw []byte) (*ParsedEmail, error) {
	entity, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("read email: %w", err)
	}

	mailHeader := &mail.Header{
		Header: entity.Header,
	}

	from, err := mailHeader.AddressList("From")
	if err != nil {
		return nil, fmt.Errorf("parse from: %w", err)
	}

	var sender string

	if len(from) > 0 {
		sender = from[0].Address

		if from[0].Name != "" {
			sender = from[0].Name + " <" + from[0].Address + ">"
		}
	}

	subject, err := mailHeader.Subject()
	if err != nil {
		return nil, fmt.Errorf("parse subject: %w", err)
	}

	textBody, err := ParseTextBody(raw)
	if err != nil {
		return nil, err
	}

	htmlBody, err := ParseHTMLBody(raw)
	if err != nil {
		return nil, err
	}

	return &ParsedEmail{
		From:    sender,
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
	}, nil
}