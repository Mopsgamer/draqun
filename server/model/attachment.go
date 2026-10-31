package model

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type Attachment struct {
	checkEmpty
	Id             uint64   `db:"id" goqu:"skipinsert"`
	MessageId      uint64   `db:"message_id"`
	AttachmentUrl  string   `db:"attachment_url"`
	AttachmentName string   `db:"attachment_name"`
	AttachmentSize int64    `db:"attachment_size"`
	AttachmentType string   `db:"attachment_type"`
	CreatedAt      TimePast `db:"created_at"`
}

var _ Model = (*Attachment)(nil)

func NewAttachment(messageId uint64, url, name, mimeType string, size int64) Attachment {
	return Attachment{
		MessageId:      messageId,
		AttachmentUrl:  url,
		AttachmentName: name,
		AttachmentType: mimeType,
		AttachmentSize: size,
		CreatedAt:      TimePast(time.Now()),
	}
}

func (a Attachment) Validate() htmx.Alert {
	if a.AttachmentUrl == "" {
		return htmx.AlertFormatMessageContent
	}
	return nil
}

func (a *Attachment) Insert() error {
	return InsertId(TableMessageAttachments, a, &a.Id)
}

func (a Attachment) Update() error {
	return Update(TableMessageAttachments, a, goqu.Ex{"id": a.Id})
}

func (a Attachment) Delete() error {
	return Delete(TableMessageAttachments, goqu.Ex{"id": a.Id})
}

func (a Attachment) IsImage() bool {
	if a.AttachmentUrl == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(a.AttachmentName))
	switch ext {
	case ".svg", ".xml", ".html", ".htm", ".xhtml", ".js", ".ts", ".docx", ".doc", ".pdf", ".json", ".css", ".zip", ".rar", ".exe", ".sh":
		return false
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico":
		return true
	}
	if strings.HasPrefix(a.AttachmentType, "image/") && !strings.Contains(a.AttachmentType, "svg") {
		return true
	}
	return false
}

func (a Attachment) IsVideo() bool {
	if a.AttachmentUrl == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(a.AttachmentName))
	switch ext {
	case ".mp4", ".webm", ".mov", ".m4v", ".mkv", ".avi", ".ogv":
		return true
	}
	if strings.HasPrefix(a.AttachmentType, "video/") {
		return true
	}
	return false
}

func (a Attachment) IsAudio() bool {
	if a.AttachmentUrl == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(a.AttachmentName))
	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac", ".opus":
		return true
	}
	if strings.HasPrefix(a.AttachmentType, "audio/") {
		return true
	}
	return false
}

func (a Attachment) IsPDF() bool {
	ext := strings.ToLower(filepath.Ext(a.AttachmentName))
	return ext == ".pdf" || strings.Contains(a.AttachmentType, "pdf")
}

func (a Attachment) IsCode() bool {
	ext := strings.ToLower(filepath.Ext(a.AttachmentName))
	switch ext {
	case ".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".rs", ".c", ".cpp", ".h", ".java", ".json", ".html", ".css", ".sql", ".sh", ".yaml", ".yml", ".md", ".xml", ".toml":
		return true
	}
	return false
}

func (a Attachment) FormattedSize() string {
	bytes := a.AttachmentSize
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024.0)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024.0*1024.0))
	}
}

func GetMessageAttachments(messageId uint64) ([]Attachment, error) {
	var attachments []Attachment
	sql, args, err := Goqu.From(TableMessageAttachments).Where(goqu.C("message_id").Eq(messageId)).Order(goqu.C("id").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}
	err = Sqlx.Select(&attachments, sql, args...)
	if err != nil {
		return nil, err
	}
	for i := range attachments {
		attachments[i].SetFullfilled()
	}
	return attachments, nil
}

type ChannelAttachment struct {
	checkEmpty
	Id        uint64   `db:"id" goqu:"skipinsert"`
	ChannelId uint64   `db:"channel_id"`
	Filename  string   `db:"filename"`
	FilePath  string   `db:"file_path"`
	FileSize  int64    `db:"file_size"`
	FileType  string   `db:"file_type"`
	CreatedAt TimePast `db:"created_at"`
}

var _ Model = (*ChannelAttachment)(nil)

func NewChannelAttachment(channelId uint64, path, name, mimeType string, size int64) ChannelAttachment {
	return ChannelAttachment{
		ChannelId: channelId,
		FilePath:  path,
		Filename:  name,
		FileType:  mimeType,
		FileSize:  size,
		CreatedAt: TimePast(time.Now()),
	}
}

func (a ChannelAttachment) Validate() htmx.Alert {
	if a.FilePath == "" {
		return htmx.AlertFormatMessageContent
	}
	return nil
}

func (a *ChannelAttachment) Insert() error {
	return InsertId(TableChannelAttachments, a, &a.Id)
}

func (a ChannelAttachment) Update() error {
	return Update(TableChannelAttachments, a, goqu.Ex{"id": a.Id})
}

func (a ChannelAttachment) Delete() error {
	return Delete(TableChannelAttachments, goqu.Ex{"id": a.Id})
}

func (a ChannelAttachment) IsImage() bool {
	if a.FilePath == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(a.Filename))
	switch ext {
	case ".svg", ".xml", ".html", ".htm", ".xhtml", ".js", ".ts", ".docx", ".doc", ".pdf", ".json", ".css", ".zip", ".rar", ".exe", ".sh":
		return false
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico":
		return true
	}
	if strings.HasPrefix(a.FileType, "image/") && !strings.Contains(a.FileType, "svg") {
		return true
	}
	return false
}

func (a ChannelAttachment) IsVideo() bool {
	if a.FilePath == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(a.Filename))
	switch ext {
	case ".mp4", ".webm", ".mov", ".m4v", ".mkv", ".avi", ".ogv":
		return true
	}
	if strings.HasPrefix(a.FileType, "video/") {
		return true
	}
	return false
}

func (a ChannelAttachment) IsAudio() bool {
	if a.FilePath == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(a.Filename))
	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac", ".opus":
		return true
	}
	if strings.HasPrefix(a.FileType, "audio/") {
		return true
	}
	return false
}

func (a ChannelAttachment) IsPDF() bool {
	ext := strings.ToLower(filepath.Ext(a.Filename))
	return ext == ".pdf" || strings.Contains(a.FileType, "pdf")
}

func (a ChannelAttachment) IsCode() bool {
	ext := strings.ToLower(filepath.Ext(a.Filename))
	switch ext {
	case ".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".rs", ".c", ".cpp", ".h", ".java", ".json", ".html", ".css", ".sql", ".sh", ".yaml", ".yml", ".md", ".xml", ".toml":
		return true
	}
	return false
}

func (a ChannelAttachment) FormattedSize() string {
	bytes := a.FileSize
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024.0)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024.0*1024.0))
	}
}

func GetChannelAttachments(channelId uint64) ([]ChannelAttachment, error) {
	var attachments []ChannelAttachment
	sql, args, err := Goqu.From(TableChannelAttachments).Where(goqu.C("channel_id").Eq(channelId)).Order(goqu.C("id").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}
	err = Sqlx.Select(&attachments, sql, args...)
	if err != nil {
		return nil, err
	}
	for i := range attachments {
		attachments[i].SetFullfilled()
	}
	return attachments, nil
}

func GetChannelAttachmentById(id uint64) (*ChannelAttachment, error) {
	var attachment ChannelAttachment
	sql, args, err := Goqu.From(TableChannelAttachments).Where(goqu.C("id").Eq(id)).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	err = Sqlx.Get(&attachment, sql, args...)
	if err != nil {
		return nil, err
	}
	attachment.SetFullfilled()
	return &attachment, nil
}
