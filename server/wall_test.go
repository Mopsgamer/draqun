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

func TestWallChannelSavingAndMultiAttachments(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Wall User", "wall_user", "wall_user@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Wall Group", "wall_group")

	// 1. Create a wall channel
	respCreateCh, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "news_wall",
			"moniker":     "News Wall",
			"description": "Initial wall text",
			"type":        "wall",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create wall channel failed: %v", err)
	}
	respCreateCh.Close()

	channels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(channels) == 0 {
		t.Fatalf("expected wall channel, got %d (err: %v)", len(channels), err)
	}
	wallCh := channels[0]
	if !wallCh.IsWall() {
		t.Fatalf("expected channel type to be wall, got %s", wallCh.Type)
	}

	// 2. GET wall view and wall edit
	respView, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/channels/%d/wall/view", group.Id, wallCh.Id))
	if err != nil {
		t.Fatalf("GET wall view failed: %v", err)
	}
	defer respView.Close()
	if respView.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for wall view, got %d", respView.StatusCode())
	}
	if !strings.Contains(respView.String(), "Initial wall text") {
		t.Errorf("expected initial wall text in view, got: %s", respView.String())
	}

	respEdit, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/channels/%d/wall/edit", group.Id, wallCh.Id))
	if err != nil {
		t.Fatalf("GET wall edit failed: %v", err)
	}
	defer respEdit.Close()
	if respEdit.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for wall edit, got %d", respEdit.StatusCode())
	}
	if !strings.Contains(respEdit.String(), "hx-encoding=\"multipart/form-data\"") {
		t.Errorf("expected hx-encoding in wall edit, got: %s", respEdit.String())
	}
	if !strings.Contains(respEdit.String(), "wall-attachment-preview") {
		t.Errorf("expected wall-attachment-preview in wall edit, got: %s", respEdit.String())
	}

	// 3. Post updated markdown content and MULTIPLE attachments
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	updatedContent := "# Welcome to the Wall\n\n- Updates item 1\n- Updates item 2"
	_ = writer.WriteField("content", updatedContent)

	// Attachment 1: doc1.txt
	f1, _ := writer.CreateFormFile("attachments", "doc1.txt")
	_, _ = f1.Write([]byte("Content of document 1"))

	// Attachment 2: doc2.png
	f2, _ := writer.CreateFormFile("attachments", "doc2.png")
	_, _ = f2.Write([]byte("\x89PNG\r\n\x1a\nfake png data"))

	_ = writer.Close()

	respSave, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetHeader("HX-Request", "true").
		SetRawBody(bodyBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/wall", group.Id, wallCh.Id))
	if err != nil {
		t.Fatalf("POST wall save failed: %v", err)
	}
	defer respSave.Close()

	if respSave.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for wall save, got %d: %s", respSave.StatusCode(), respSave.String())
	}
	saveHtml := respSave.String()
	if !strings.Contains(saveHtml, "Welcome to the Wall") {
		t.Errorf("expected updated content in rendered wall view, got: %s", saveHtml)
	}
	if !strings.Contains(saveHtml, "doc1.txt") || !strings.Contains(saveHtml, "doc2.png") {
		t.Errorf("expected multi-attachments doc1.txt and doc2.png in rendered wall view, got: %s", saveHtml)
	}

	// Check DB
	chUpdated, err := model.NewChannelFromId(wallCh.Id)
	if err != nil {
		t.Fatalf("failed to fetch channel: %v", err)
	}
	if string(chUpdated.Description) != updatedContent {
		t.Errorf("expected description to be %q, got %q", updatedContent, chUpdated.Description)
	}

	attachments, err := model.GetChannelAttachments(wallCh.Id)
	if err != nil {
		t.Fatalf("failed to fetch channel attachments: %v", err)
	}
	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments in DB, got %d", len(attachments))
	}

	// 4. Delete one attachment
	delAttId := attachments[0].Id
	respDel, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Delete(fmt.Sprintf("/groups/%d/channels/%d/wall/attachments/%d", group.Id, wallCh.Id, delAttId))
	if err != nil {
		t.Fatalf("DELETE wall attachment failed: %v", err)
	}
	defer respDel.Close()
	if respDel.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for attachment delete, got %d", respDel.StatusCode())
	}

	remAttachments, _ := model.GetChannelAttachments(wallCh.Id)
	if len(remAttachments) != 1 {
		t.Errorf("expected 1 remaining attachment, got %d", len(remAttachments))
	}

	// 5. Subsequent edit with a third attachment should retain existing attachments
	bodyBuf2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(bodyBuf2)
	_ = writer2.WriteField("content", "# Wall Edit 2\n\nPreserve previous attachments.")
	f3, _ := writer2.CreateFormFile("attachments", "doc3.pdf")
	_, _ = f3.Write([]byte("%PDF-1.4 fake pdf data"))
	_ = writer2.Close()

	respSave2, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", writer2.FormDataContentType()).
		SetHeader("HX-Request", "true").
		SetRawBody(bodyBuf2.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/wall", group.Id, wallCh.Id))
	if err != nil {
		t.Fatalf("POST wall save 2 failed: %v", err)
	}
	defer respSave2.Close()
	if respSave2.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for wall save 2, got %d", respSave2.StatusCode())
	}

	finalAttachments, err := model.GetChannelAttachments(wallCh.Id)
	if err != nil {
		t.Fatalf("failed to fetch channel attachments after save 2: %v", err)
	}
	// Should have 1 remaining from before + 1 newly added = 2 attachments
	if len(finalAttachments) != 2 {
		t.Fatalf("expected 2 attachments after save 2, got %d", len(finalAttachments))
	}
}
