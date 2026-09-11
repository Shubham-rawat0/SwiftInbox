package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/parser"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

type MessageHandler struct {
	queries *postgres.Queries
}

type MessagePreview struct {
	ID        uuid.UUID `json:"id"`
	Sender    string    `json:"sender"`
	Subject   string    `json:"subject"`
	Preview   string    `json:"preview"`
	CreatedAt time.Time `json:"createdAt"`
}

type MessageResponse struct {
	Address      string           `json:"address"`
	Messages     []MessagePreview `json:"messages"`
	MessageCount int              `json:"messageCount"`
}

type ParsedMessage struct {
	ID         uuid.UUID       `json:"id"`
	From       string       `json:"from"`
	Subject    string       `json:"subject"`
	Body       []byte      `json:"body"`
	CreatedAt  time.Time    `json:"createdAt"`
	Mailbox    string       `json:"mailbox"`
	ParsedData ParsedData   `json:"parsedData"`
}

type ParsedData struct {
	Subject     string       `json:"subject"`
	From        string       `json:"from"`
	Text        string       `json:"text"`
	HTML        string       `json:"html"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	ContentID   string `json:"contentId"`
	Index       int    `json:"index"`
}

func NewMessageHandler(q *postgres.Queries) *MessageHandler {
	return &MessageHandler{
		queries: q,
	}
}

func (m *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("address")
	address := utils.NormalizeAddress(identifier)

	var data []postgres.GetMessagesRow
	var err error
	
	data, err = m.queries.GetMessages(r.Context(), address)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(
				w,
				http.StatusNotFound,
				errors.New("mailbox not found"),
			)
			return
		}

		WriteError(
			w,
			http.StatusInternalServerError,
			errors.New("failed to fetch messages"),
		)
		return
	}

	messages := make([]MessagePreview, 0, len(data))

	for _, msg := range data {
		preview, err := parser.ParseTextBody(msg.Raw)

		if err != nil {
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

	response := MessageResponse{
		Address:      address,
		Messages:     messages,
		MessageCount: len(messages),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (m *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id:=r.PathValue("id")
	if id==""{
		WriteError(w,http.StatusBadRequest,errors.New("need a message id"))
		return
	}
	Id,err:=uuid.Parse(id)
	if err!=nil{
		WriteError(w,http.StatusInternalServerError,errors.New("error parsing id"))
		return
	}
	data,err:=m.queries.GetMessage(r.Context(),Id,)
	if err!=nil{
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(
				w,
				http.StatusNotFound,
				errors.New("message not found"),
			)
			return
		}
		WriteError(w,http.StatusInternalServerError,err)
	}

	parsedMessage:=ParsedMessage{
		ID: data.ID,
		From      :data.Sender,
		Subject   :data.Subject.String,
		Body      :data.Raw,
		CreatedAt  :data.CreatedAt,
		Mailbox    :data.Address,	}

	parseBody,err:=parser.ParseEmail(data.Raw)
	if err!=nil{
		WriteError(w,http.StatusInternalServerError,err)
	}
	
	attachments,err:=parser.ParseAttachments(data.Raw)
	if err!=nil{
		WriteError(w,http.StatusInternalServerError,err)
	}

	attachment := make([]Attachment, 0)

	for i:=range(len(attachments)){
		attachment=append(attachment, Attachment{
			Filename: attachments[i].Filename,
			ContentType: attachments[i].ContentType,
			ContentID: attachments[i].ContentID,
			Size: attachments[i].Size,
			Index: attachments[i].Index,
		})
	}
	parsedData:= ParsedData{
		From: parseBody.From,
		Subject: parseBody.Subject,
		Text: parseBody.Text,
		HTML: parseBody.HTML,
		Attachments: attachment,
	 }

	 parsedMessage.ParsedData=parsedData

	 WriteJSON(w,http.StatusOK,parsedMessage)
	
}

func (m *MessageHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	indexStr := r.PathValue("index")

	if id == "" {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("need a message id"),
		)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("invalid attachment index"),
		)
		return
	}

	if index < 0 {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("index should be greater than or equal to 0"),
		)
		return
	}

	messageID, err := uuid.Parse(id)
	if err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("invalid message id"),
		)
		return
	}

	data, err := m.queries.GetMessage(r.Context(), messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(
				w,
				http.StatusNotFound,
				errors.New("message not found"),
			)
			return
		}

		WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	attachments, err := parser.ParseAttachments(data.Raw)
	if err != nil {
		WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	if index >= len(attachments) {
		WriteError(
			w,
			http.StatusNotFound,
			errors.New("attachment not found"),
		)
		return
	}

	attachment := attachments[index]

	if attachment.ContentType != "" {
		w.Header().Set("Content-Type", attachment.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if attachment.Filename != "" && attachment.ContentID == "" {
		w.Header().Set(
			"Content-Disposition",
			fmt.Sprintf(`attachment; filename="%s"`, attachment.Filename),
		)
	}

	if attachment.ContentID != "" {
		w.Header().Set(
			"Cache-Control",
			"public, max-age=3600",
		)
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(attachment.Data); err != nil {
		log.Printf("failed to write attachment: %v", err)
	}
}