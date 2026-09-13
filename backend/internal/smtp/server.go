package smtp

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/parser"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	gosmtp "github.com/emersion/go-smtp"
	"github.com/google/uuid"
)

type SMTPServer struct {
	queries *postgres.Queries
}

func NewSMTPServer(queries *postgres.Queries) *SMTPServer {
	return &SMTPServer{
		queries: queries,
	}
}

type Session struct {
	queries *postgres.Queries

	from       string
	recipients []string
}

func (s *Session) Mail(from string, opts *gosmtp.MailOptions) error {
	s.from = utils.NormalizeAddress(from)
	s.recipients = nil

	log.Printf("[MAIL FROM] %s", s.from)

	return nil
}

func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	rcpt := utils.NormalizeAddress(to)

	log.Printf(
		"[RCPT TO] raw=%s normalized=%s",
		to,
		rcpt,
	)

	if !utils.IsOurDomain(rcpt) {
		domain := utils.ExtractDomain(rcpt)

		return fmt.Errorf(
			"relay denied for domain %s",
			domain,
		)
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := s.queries.UpsertMailbox(
		context.Background(),
		postgres.UpsertMailboxParams{
			ID:        uuid.New(),
			Address:   rcpt,
			ExpiresAt: expiresAt,
		},
	)

	if err != nil {
		log.Printf(
			"[RCPT ERROR] address=%s error=%v",
			rcpt,
			err,
		)

		return err
	}

	log.Printf("[RCPT ACCEPTED] %s", rcpt)

	s.recipients = append(s.recipients, rcpt)

	return nil
}

func (s *Session) Data(r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		log.Printf(
			"[MESSAGE ERROR] failed reading SMTP DATA: %v",
			err,
		)

		return err
	}

	if len(s.recipients) == 0 {
		return fmt.Errorf("no recipients")
	}

	parsed, err := parser.ParseEmail(raw)
	if err != nil {
		log.Printf(
			"[MESSAGE ERROR] failed parsing email: %v",
			err,
		)

		return err
	}

	if parsed.From == "" {
		parsed.From = "unknown"
	}

	if parsed.Subject == "" {
		parsed.Subject = "(No Subject)"
	}

	log.Printf(
		"[MESSAGE] to=%v from=%s subject=%s",
		s.recipients,
		parsed.From,
		parsed.Subject,
	)

	for _, recipient := range s.recipients {
		if err := s.storeMessage(
			recipient,
			parsed,
			raw,
		); err != nil {

			log.Printf(
				"[MESSAGE ERROR] recipient=%s error=%v",
				recipient,
				err,
			)

			return err
		}

		log.Printf(
			"[MESSAGE STORED] %s",
			recipient,
		)
	}

	return nil
}

func (s *Session) storeMessage(recipient string, parsed *parser.ParsedEmail, raw []byte) error {

	ctx := context.Background()
	data, err := s.queries.GetMailboxId(ctx, recipient)
	id := uuid.New()
	if err != nil {
		return err
	}

	createdBy, err := s.queries.GetMailboxCreatedBy(ctx, data)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if createdBy.Valid {
		expiresAt = time.Now().Add(30 * 24 * time.Hour)
	}
	_, err = s.queries.CreateMessage(
		ctx,
		postgres.CreateMessageParams{
			ID:        id,
			Address:   recipient,
			MailboxID: data,
			Sender:    parsed.From,
			Subject: sql.NullString{
				String: parsed.Subject,
				Valid:  parsed.Subject != "",
			},
			Raw:       raw,
			ExpiresAt: expiresAt,
		},
	)

	return err
}

type Backend struct {
	queries *postgres.Queries
}

func (b *Backend) NewSession(conn *gosmtp.Conn) (gosmtp.Session, error) {

	return &Session{
		queries: b.queries,
	}, nil
}

func (s *Session) Logout() error {
	log.Printf("[SMTP LOGOUT] from=%s recipients=%v", s.from, s.recipients)

	return nil
}

func (s *Session) Reset() {
	s.from = ""
	s.recipients = nil
}

func (s *SMTPServer) Start() error {
	port := 2525

	if value := os.Getenv("SMTP_PORT"); value != "" {
		parsedPort, err := strconv.Atoi(value)

		if err != nil {
			return fmt.Errorf(
				"invalid SMTP_PORT %q: %w",
				value,
				err,
			)
		}

		port = parsedPort
	}

	server := gosmtp.NewServer(
		&Backend{
			queries: s.queries,
		},
	)

	server.Addr = fmt.Sprintf(
		"0.0.0.0:%d",
		port,
	)

	server.Domain = utils.GetAllowedDomain()

	server.AllowInsecureAuth = false

	log.Printf(
		"SMTP server listening on port %d , domain %v",
		port, server.Domain,
	)

	return server.ListenAndServe()
}
