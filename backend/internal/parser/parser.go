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
		Index       int
		Inline      bool
		Data        []byte
}

func ParseTextBody(raw []byte) (string, error) {
	entity, err := message.Read(bytes.NewReader(normalizeFinalBoundary(raw)))
	if err != nil {
		return "", fmt.Errorf("read email: %w", err)
	}

	var plain, html string
	if err := walk(entity, func(contentType string, params map[string]string, header *message.Header, data []byte) error {
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
	entity, err := message.Read(bytes.NewReader(normalizeFinalBoundary(raw)))
	if err != nil {
		return "", fmt.Errorf("read email: %w", err)
	}

	var html string
	if err := walk(entity, func(contentType string, params map[string]string, header *message.Header, data []byte) error {
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
	entity, err := message.Read(bytes.NewReader(normalizeFinalBoundary(raw)))
	if err != nil {
		return nil, fmt.Errorf("read email: %w", err)
	}

	var attachments []Attachment
	if err := walk(entity, func(contentType string, params map[string]string, header *message.Header, data []byte) error {
		isBody := contentType == "text/plain" || contentType == "text/html"

		disp, dispParams, _ := header.ContentDisposition()
		filename := params["name"]
		if filename == "" {
			filename = dispParams["filename"]
		}

		contentID := strings.Trim(header.Get("Content-Id"), "<> ")
		if contentID == "" {
			contentID = strings.Trim(params["content-id"], "<> ")
		}

		// Only skip parts that are the email body itself (no explicit
		// filename, name param, content-id or attachment disposition).
		// Named text/plain (or text/html) parts ARE attachments.
		if isBody && filename == "" && contentID == "" && !strings.EqualFold(disp, "attachment") {
			return nil
		}

		inline := strings.EqualFold(disp, "inline")
		if !inline && contentID != "" && !strings.EqualFold(disp, "attachment") {
			inline = true
		}

		attachments = append(attachments, Attachment{
			Filename:    filename,
			ContentType: contentType,
			ContentID:   contentID,
			Inline:      inline,
			Size:        len(data),
			Index:       len(attachments),
			Data:        data,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return attachments, nil
}

func normalizeFinalBoundary(raw []byte) []byte {
	if !bytes.HasSuffix(raw, []byte("\r\n")) {
		return raw
	}

	headerEnd := bytes.Index(raw, []byte("\n\n"))
	if headerEnd < 0 {
		return raw
	}

	body := raw[headerEnd+2:]
	firstLineEnd := bytes.IndexByte(body, '\n')
	if firstLineEnd < 0 || (firstLineEnd > 0 && body[firstLineEnd-1] == '\r') {
		return raw
	}

	normalized := make([]byte, len(raw)-1)
	copy(normalized, raw[:len(raw)-2])
	normalized[len(normalized)-1] = '\n'
	return normalized
}

func walk(entity *message.Entity, fn func(contentType string, params map[string]string, header *message.Header, data []byte) error) error {
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

	return fn(contentType, params, &entity.Header, data)
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
