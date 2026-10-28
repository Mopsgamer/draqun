package internal

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestMultiAttachmentAndMarkdown(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Multi User", "multi_user", "multi_user@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Multi Group", "multi_group")

	// Create channel
	respCreateCh, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "general",
			"moniker":     "General Channel",
			"description": "General channel",
			"type":        "text",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create channel failed: %v", err)
	}
	respCreateCh.Close()

	channels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(channels) == 0 {
		t.Fatalf("expected channel, got %d (err: %v)", len(channels), err)
	}
	ch := channels[0]

	// 1. Post a message with Markdown text and MULTIPLE attachments
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	_ = writer.WriteField("content", "Here is a message with **bold** text\n\n> [!NOTE]\n> A note callout")

	// File 1: photo.jpg
	f1, _ := writer.CreateFormFile("attachments", "photo.jpg")
	_, _ = f1.Write([]byte("\xff\xd8\xff\xe0\x00\x10JFIF"))

	// File 2: clip.mp4
	f2, _ := writer.CreateFormFile("attachments", "clip.mp4")
	_, _ = f2.Write([]byte("fake mp4 content"))

	// File 3: doc.pdf
	f3, _ := writer.CreateFormFile("attachments", "doc.pdf")
	_, _ = f3.Write([]byte("%PDF-1.4 fake pdf content"))

	_ = writer.Close()

	resp, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetHeader("HX-Request", "true").
		SetRawBody(bodyBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("post multi-attachment message failed: %v", err)
	}
	resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for multi-attachment message, got %d", resp.StatusCode())
	}

	// Fetch message from DB
	msgs := ch.MessagesPage(1, 10)
	if len(msgs) == 0 {
		t.Fatalf("expected message in DB")
	}
	msg := msgs[0]

	attachments := msg.GetAttachments()
	if len(attachments) != 3 {
		t.Fatalf("expected 3 attachments, got %d", len(attachments))
	}

	// Verify attachment types
	if !attachments[0].IsImage() {
		t.Errorf("expected photo.jpg to be Image")
	}
	if !attachments[1].IsVideo() {
		t.Errorf("expected clip.mp4 to be Video")
	}
	if !attachments[2].IsPDF() {
		t.Errorf("expected doc.pdf to be PDF")
	}

	// 2. Fetch rendered HTML and verify markdown + multi-attachment HTML elements
	respPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET channel messages page failed: %v", err)
	}
	htmlStr := respPage.String()
	respPage.Close()

	if !strings.Contains(htmlStr, "<strong>bold</strong>") {
		t.Errorf("expected markdown rendered bold text, got:\n%s", htmlStr)
	}
	if !strings.Contains(htmlStr, "callout-note") {
		t.Errorf("expected note callout rendered in HTML, got:\n%s", htmlStr)
	}
	if !strings.Contains(htmlStr, "photo.jpg") || !strings.Contains(htmlStr, "clip.mp4") || !strings.Contains(htmlStr, "doc.pdf") {
		t.Errorf("expected all 3 attachments in rendered HTML")
	}
	if !strings.Contains(htmlStr, "<video") {
		t.Errorf("expected video element for clip.mp4")
	}
}

func TestAttachmentModelHelpers(t *testing.T) {
	img := model.NewAttachment(1, "/uploads/file.png", "image.png", "image/png", 2048)
	if !img.IsImage() {
		t.Errorf("expected IsImage true for image.png")
	}
	if img.FormattedSize() != "2.0 KB" {
		t.Errorf("expected '2.0 KB', got '%s'", img.FormattedSize())
	}

	aud := model.NewAttachment(1, "/uploads/file.mp3", "song.mp3", "audio/mpeg", 1048576)
	if !aud.IsAudio() {
		t.Errorf("expected IsAudio true for song.mp3")
	}
	if aud.FormattedSize() != "1.0 MB" {
		t.Errorf("expected '1.0 MB', got '%s'", aud.FormattedSize())
	}

	code := model.NewAttachment(1, "/uploads/file.go", "main.go", "text/plain", 100)
	if !code.IsCode() {
		t.Errorf("expected IsCode true for main.go")
	}
	if code.FormattedSize() != "100 B" {
		t.Errorf("expected '100 B', got '%s'", code.FormattedSize())
	}
}
