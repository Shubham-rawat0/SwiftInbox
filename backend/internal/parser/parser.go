package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-message"
)

type ParsedEmail struct {
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func ParseRawMail(raw []byte) (*ParsedEmail, error) {
	parsed := &ParsedEmail{}

	entity, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("read email: %w", err)
	}

	if err := parseEntity(entity, parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func parseEntity(entity *message.Entity,parsed *ParsedEmail,) error {
	contentType, _, err := entity.Header.ContentType()
	if err != nil {
		return fmt.Errorf("parse content type: %w", err)
	}

	if strings.HasPrefix(contentType, "multipart/") {
		return parseMultipart(entity, parsed)
	}

	return parseBody(entity, parsed)
}

func parseMultipart(entity *message.Entity, parsed *ParsedEmail,) error {
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

		if err := parseEntity(part, parsed); err != nil {
			return err
		}
	}

	return nil
}

func parseBody(entity *message.Entity, parsed *ParsedEmail,) error {
	contentType, params, err := entity.Header.ContentType()
	if err != nil {
		return fmt.Errorf("parse content type: %w", err)
	}

	data, err := io.ReadAll(entity.Body)
	if err != nil {
		return fmt.Errorf("read MIME body: %w", err)
	}

	switch contentType {

	case "text/plain":
		parsed.TextBody = string(data)

	case "text/html":
		parsed.HTMLBody = string(data)

	default:
		filename := params["name"]

		parsed.Attachments = append(
			parsed.Attachments,
			Attachment{
				Filename:    filename,
				ContentType: contentType,
				Data:        data,
			},
		)
	}

	return nil
}