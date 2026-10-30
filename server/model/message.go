package model

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type Message struct {
	checkEmpty
	Id             uint64         `db:"id" goqu:"skipinsert"`
	GroupId        uint64         `db:"group_id"`
	ChannelId      uint64         `db:"channel_id"`
	AuthorId       uint64         `db:"author_id"`
	Content        MessageContent `db:"content"`
	AttachmentUrl  string         `db:"attachment_url"`
	AttachmentName string         `db:"attachment_name"`
	AttachmentSize int64          `db:"attachment_size"`
	AttachmentType string         `db:"attachment_type"`
	CreatedAt      TimePast       `db:"created_at"`
}

var _ Model = (*Message)(nil)

func NewMessageFilled(channelId, groupId, userId uint64, content string) Message {
	return Message{
		GroupId:   groupId,
		ChannelId: channelId,
		AuthorId:  userId,
		Content:   MessageContent(strings.TrimSpace(content)),
		CreatedAt: TimePast(time.Now()),
	}
}

func (message Message) Validate() htmx.Alert {
	if message.AttachmentUrl == "" {
		if !message.Content.IsValid() {
			return htmx.AlertFormatMessageContent
		}
	} else if len(message.Content) > environment.ChatMessageMaxLength {
		return htmx.AlertFormatMessageContent
	}

	if !message.CreatedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (message Message) IsImageAttachment() bool {
	if message.AttachmentUrl == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(message.AttachmentName))
	// Exclude SVG and vector/scriptable/executable formats from inline preview
	switch ext {
	case ".svg", ".xml", ".html", ".htm", ".xhtml", ".js", ".ts", ".docx", ".doc", ".pdf", ".json", ".css", ".zip", ".rar", ".exe", ".sh":
		return false
	}
	// Restrict allowed inline raster image formats
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico":
		return true
	}
	if strings.HasPrefix(message.AttachmentType, "image/") && !strings.Contains(message.AttachmentType, "svg") {
		return true
	}
	return false
}

func (message Message) FormattedSize() string {
	bytes := message.AttachmentSize
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024.0)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024.0*1024.0))
	}
}

func (message *Message) Insert() error {
	return InsertId(TableMessages, message, &message.Id)
}

func (message Message) Update() error {
	return Update(TableMessages, message, goqu.Ex{"id": message.Id})
}

func NewMessageFromId(id uint64) (Message, error) {
	message := Message{}
	if id == 0 {
		return message, nil
	}
	sql, args, err := Goqu.From(TableMessages).Where(goqu.Ex{"id": id}).Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message, err
	}
	err = Sqlx.Get(&message, sql, args...)
	if err == nil {
		message.SetFullfilled()
	}
	return message, err
}

func (message Message) Delete() error {
	return Delete(TableMessages, goqu.Ex{"id": message.Id})
}

func (message *Message) Author() User {
	user, _ := NewUserFromId(message.AuthorId)
	return user
}

func (message Message) GetAttachments() []Attachment {
	if message.Id != 0 {
		attachments, err := GetMessageAttachments(message.Id)
		if err == nil {
			return attachments
		}
	}
	return nil
}
