package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/parser"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrMailboxNotFound           = errors.New("mailbox not found")
	ErrMailboxAlreadyExists       = errors.New("email address already exists")
	ErrWrongDomain               = errors.New("wrong domain")
	ErrInvalidAddress            = errors.New("invalid email address")
	ErrForbiddenDeveloperMailbox = errors.New("this email belongs to a developer mailbox; please choose another mailbox")
	ErrInvalidExpiry             = errors.New("expiresAt must be in the future")
	ErrDeveloperOnlyExpiry       = errors.New("only developers can set expiresAt")
	ErrMessageNotFound           = errors.New("message not found")
	ErrAttachmentNotFound        = errors.New("attachment not found")
	ErrInvalidAttachmentIndex    = errors.New("attachment index out of bounds")
	ErrUsernameRequired          = errors.New("username is required")
	ErrAddressRequired           = errors.New("address is required")
	ErrUnauthorized              = errors.New("unauthorized")
)


type MailboxResult struct {
	ID        uuid.UUID     `json:"id,omitempty"`
	Address   string        `json:"address"`
	CreatedAt time.Time     `json:"createdAt"`
	ExpiresAt time.Time     `json:"expiresAt"`
	CreatedBy uuid.NullUUID `json:"createdBy"`
}

type DeletedMailboxResult struct {
	ID      uuid.UUID `json:"id"`
	Address string    `json:"address"`
}

type MessagePreview struct {
	ID        uuid.UUID `json:"id"`
	Sender    string    `json:"sender"`
	Subject   string    `json:"subject"`
	Preview   string    `json:"preview"`
	CreatedAt time.Time `json:"createdAt"`
}

type MessagesResult struct {
	Address      string           `json:"address"`
	Messages     []MessagePreview `json:"messages"`
	MessageCount int              `json:"messageCount"`
}

type AttachmentInfo struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	ContentID   string `json:"contentId"`
	Index       int    `json:"index"`
}

type ParsedData struct {
	Subject     string           `json:"subject"`
	From        string           `json:"from"`
	Text        string           `json:"text"`
	HTML        string           `json:"html"`
	Attachments []AttachmentInfo `json:"attachments"`
}

type MessageDetailResult struct {
	ID         uuid.UUID  `json:"id"`
	From       string     `json:"from"`
	Subject    string     `json:"subject"`
	Body       []byte     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	Mailbox    string     `json:"mailbox"`
	ParsedData ParsedData `json:"parsedData"`
}

type AttachmentResult struct {
	Filename    string
	ContentType string
	ContentID   string
	Data        []byte
}

type MailboxService struct {
	queries *postgres.Queries
}

func NewMailboxService(q *postgres.Queries) *MailboxService {
	return &MailboxService{
		queries: q,
	}
}

func (s *MailboxService) calculateExpiry(requested *time.Time, isDeveloper bool) (time.Time, error) {
	if requested == nil {
		if isDeveloper {
			return time.Now().Add(30 * 24 * time.Hour).UTC(), nil
		}
		return time.Now().Add(24 * time.Hour).UTC(), nil
	}

	if !isDeveloper {
		return time.Time{}, ErrDeveloperOnlyExpiry
	}

	if !requested.After(time.Now()) {
		return time.Time{}, ErrInvalidExpiry
	}

	return requested.UTC(), nil
}

func (s *MailboxService) CreateMailbox(ctx context.Context, requestedAddress string, devID *uuid.UUID, requestedExpiry *time.Time) (*MailboxResult, error) {
	if requestedAddress == "" {
		return nil, ErrAddressRequired
	}

	address := utils.NormalizeAddress(requestedAddress)
	if !utils.IsOurDomain(address) {
		return nil, ErrWrongDomain
	}

	isDeveloper := devID != nil
	expiresAt, err := s.calculateExpiry(requestedExpiry, isDeveloper)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	createdByValue := uuid.NullUUID{Valid: false}
	if isDeveloper {
		createdByValue = uuid.NullUUID{UUID: *devID, Valid: true}
	}

	mb, err := s.queries.CreateEmailAddress(ctx, postgres.CreateEmailAddressParams{
		ID:        id,
		Address:   address,
		ExpiresAt: expiresAt,
		CreatedBy: createdByValue,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMailboxNotFound
		}
		return nil, err
	}

	return &MailboxResult{
		ID:        id,
		Address:   mb.Address,
		CreatedAt: mb.CreatedAt,
		ExpiresAt: mb.ExpiresAt,
		CreatedBy: mb.CreatedBy,
	}, nil
}

func (s *MailboxService) CreateCustomMailbox(ctx context.Context, username string, devID *uuid.UUID, requestedExpiry *time.Time) (*MailboxResult, error) {
	if username == "" {
		return nil, ErrUsernameRequired
	}

	address, err := utils.MakeCustomAddress(username)
	if err != nil {
		return nil, err
	}
	if address == "" {
		return nil, ErrInvalidAddress
	}

	isDeveloper := devID != nil
	expiresAt, err := s.calculateExpiry(requestedExpiry, isDeveloper)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	createdByValue := uuid.NullUUID{Valid: false}
	if isDeveloper {
		createdByValue = uuid.NullUUID{UUID: *devID, Valid: true}
	}

	data, err := s.queries.CreateCustomEmailAddress(ctx, postgres.CreateCustomEmailAddressParams{
		ID:        id,
		Address:   address,
		ExpiresAt: expiresAt,
		CreatedBy: createdByValue,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMailboxAlreadyExists
		}
		return nil, err
	}

	return &MailboxResult{
		ID:        id,
		Address:   data.Address,
		CreatedAt: data.CreatedAt,
		ExpiresAt: data.ExpiresAt,
		CreatedBy: data.CreatedBy,
	}, nil
}

func (s *MailboxService) ListDeveloperMailboxes(ctx context.Context, devID uuid.UUID) ([]MailboxResult, error) {
	rows, err := s.queries.ListDeveloperMailboxes(ctx, uuid.NullUUID{UUID: devID, Valid: true})
	if err != nil {
		return nil, err
	}

	results := make([]MailboxResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, MailboxResult{
			ID:        r.ID,
			Address:   r.Address,
			CreatedAt: r.CreatedAt,
			ExpiresAt: r.ExpiresAt,
			CreatedBy: uuid.NullUUID{UUID: devID, Valid: true},
		})
	}
	return results, nil
}

func (s *MailboxService) GetDeveloperMailbox(ctx context.Context, devID uuid.UUID, mailboxID uuid.UUID) (*MailboxResult, error) {
	row, err := s.queries.GetDeveloperMailboxByID(ctx, postgres.GetDeveloperMailboxByIDParams{
		ID:        mailboxID,
		CreatedBy: uuid.NullUUID{UUID: devID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMailboxNotFound
		}
		return nil, err
	}

	return &MailboxResult{
		ID:        row.ID,
		Address:   row.Address,
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
		CreatedBy: row.CreatedBy,
	}, nil
}

func (s *MailboxService) DeleteDeveloperMailbox(ctx context.Context, devID uuid.UUID, identifier string) (*DeletedMailboxResult, error) {
	if identifier == "" {
		return nil, ErrAddressRequired
	}

	if parsedID, err := uuid.Parse(identifier); err == nil {
		row, err := s.queries.DeleteDeveloperMailboxByID(ctx, postgres.DeleteDeveloperMailboxByIDParams{
			ID:        parsedID,
			CreatedBy: uuid.NullUUID{UUID: devID, Valid: true},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrMailboxNotFound
			}
			return nil, err
		}
		return &DeletedMailboxResult{ID: row.ID, Address: row.Address}, nil
	}

	address := utils.NormalizeAddress(identifier)
	if !utils.IsOurDomain(address) {
		return nil, ErrWrongDomain
	}

	row, err := s.queries.DeleteMailbox(ctx, postgres.DeleteMailboxParams{
		Address:   address,
		CreatedBy: uuid.NullUUID{UUID: devID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMailboxNotFound
		}
		return nil, err
	}

	return &DeletedMailboxResult{ID: row.ID, Address: row.Address}, nil
}

func (s *MailboxService) VerifyMailboxAccess(ctx context.Context, address string, devID *uuid.UUID) error {
	// A developer may access their own developer mailboxes and any public
	// mailbox. A public (unauthenticated) caller may only access public
	// mailboxes. In both cases the existence of another developer's mailbox
	// is reported as forbidden rather than leaking its presence via a
	// distinct error.
	var err error
	if devID != nil {
		_, err = s.queries.GetDeveloperMailboxId(ctx, postgres.GetDeveloperMailboxIdParams{
			Address:   address,
			CreatedBy: uuid.NullUUID{UUID: *devID, Valid: true},
		})
	} else {
		_, err = s.queries.GetPublicMailboxId(ctx, address)
	}

	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if devID != nil {
		_, err = s.queries.GetPublicMailboxId(ctx, address)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	if _, mailboxErr := s.queries.GetMailboxId(ctx, address); mailboxErr == nil {
		return ErrForbiddenDeveloperMailbox
	}

	return ErrMailboxNotFound
}

func (s *MailboxService) GetMessages(ctx context.Context, identifier string, devID *uuid.UUID) (*MessagesResult, error) {
	address := utils.NormalizeAddress(identifier)

	if err := s.VerifyMailboxAccess(ctx, address, devID); err != nil {
		return nil, err
	}

	data, err := s.queries.GetMessages(ctx, address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMailboxNotFound
		}
		return nil, err
	}

	messages := make([]MessagePreview, 0, len(data))
	for _, msg := range data {
		preview, parseErr := parser.ParseTextBody(msg.Raw)
		if parseErr != nil {
			preview = "Unable to load preview"
		}
		preview = parser.MakePreview(preview, 150)

		messages = append(messages, MessagePreview{
			ID:        msg.ID,
			Sender:    msg.Sender,
			Subject:   msg.Subject.String,
			Preview:   preview,
			CreatedAt: msg.CreatedAt,
		})
	}

	return &MessagesResult{
		Address:      address,
		Messages:     messages,
		MessageCount: len(messages),
	}, nil
}

func (s *MailboxService) GetMessage(ctx context.Context, messageID uuid.UUID, devID *uuid.UUID) (*MessageDetailResult, error) {
	data, err := s.queries.GetMessage(ctx, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}

	if err := s.VerifyMailboxAccess(ctx, data.Address, devID); err != nil {
		return nil, err
	}

	parseBody, _ := parser.ParseEmail(data.Raw)
	attachments, _ := parser.ParseAttachments(data.Raw)

	attachmentInfos := make([]AttachmentInfo, 0, len(attachments))
	for _, att := range attachments {
		attachmentInfos = append(attachmentInfos, AttachmentInfo{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			ContentID:   att.ContentID,
			Size:        att.Size,
			Index:       att.Index,
		})
	}

	parsedData := ParsedData{
		From:        parseBody.From,
		Subject:     parseBody.Subject,
		Text:        parseBody.Text,
		HTML:        parseBody.HTML,
		Attachments: attachmentInfos,
	}

	return &MessageDetailResult{
		ID:         data.ID,
		From:       data.Sender,
		Subject:    data.Subject.String,
		Body:       data.Raw,
		CreatedAt:  data.CreatedAt,
		Mailbox:    data.Address,
		ParsedData: parsedData,
	}, nil
}

func (s *MailboxService) GetAttachment(ctx context.Context, messageID uuid.UUID, index int, devID *uuid.UUID) (*AttachmentResult, error) {
	if index < 0 {
		return nil, ErrInvalidAttachmentIndex
	}

	data, err := s.queries.GetMessage(ctx, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}

	if err := s.VerifyMailboxAccess(ctx, data.Address, devID); err != nil {
		return nil, err
	}

	attachments, err := parser.ParseAttachments(data.Raw)
	if err != nil {
		return nil, err
	}

	if index >= len(attachments) {
		return nil, ErrAttachmentNotFound
	}

	att := attachments[index]
	contentType := att.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &AttachmentResult{
		Filename:    att.Filename,
		ContentType: contentType,
		ContentID:   att.ContentID,
		Data:        att.Data,
	}, nil
}
