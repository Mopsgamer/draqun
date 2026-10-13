package internal

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestFileSharingFunctionality(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "File User", "file_user", "file_user@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "File Sharing Group", "files_group")

	// Create main text channel
	respCreateMain, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "general",
			"moniker":     "General Channel",
			"description": "Channel for files",
			"type":        "text",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create channel failed: %v", err)
	}
	respCreateMain.Close()

	channels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(channels) == 0 {
		t.Fatalf("expected channel, got %d (err: %v)", len(channels), err)
	}
	ch := channels[0]

	// 1. Send message with text content AND text file attachment (.txt)
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	_ = writer.WriteField("content", "Here is a document for you.")
	filePart, err := writer.CreateFormFile("attachment", "notes.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, _ = filePart.Write([]byte("Document content inside notes.txt"))
	_ = writer.Close()

	respDocMsg, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetRawBody(bodyBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("post message with attachment failed: %v", err)
	}
	respDocMsg.Close()

	if respDocMsg.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for message with file, got %d", respDocMsg.StatusCode())
	}

	// Verify message in database
	msgs := ch.MessagesPage(1, 10)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in channel, got %d", len(msgs))
	}
	docMsg := msgs[0]
	if docMsg.Content != "Here is a document for you." {
		t.Errorf("expected content 'Here is a document for you.', got '%s'", docMsg.Content)
	}
	if docMsg.AttachmentName != "notes.txt" {
		t.Errorf("expected AttachmentName 'notes.txt', got '%s'", docMsg.AttachmentName)
	}
	if !strings.HasPrefix(docMsg.AttachmentUrl, "/uploads/") {
		t.Errorf("expected AttachmentUrl starting with '/uploads/', got '%s'", docMsg.AttachmentUrl)
	}
	if docMsg.AttachmentSize == 0 {
		t.Errorf("expected AttachmentSize > 0, got %d", docMsg.AttachmentSize)
	}

	// Check file on disk
	diskPath := strings.TrimPrefix(docMsg.AttachmentUrl, "/")
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		t.Errorf("expected file on disk at '%s', but not found", diskPath)
	}

	// Test security headers when fetching uploaded file
	baseURL := fmt.Sprintf("http://%s", ts.Listener.Addr().String())
	httpResp, err := http.Get(baseURL + docMsg.AttachmentUrl)
	if err != nil {
		t.Fatalf("http.Get uploaded file failed: %v", err)
	}
	defer httpResp.Body.Close()

	if nosniff := httpResp.Header.Get("X-Content-Type-Options"); nosniff != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got '%s'", nosniff)
	}
	if csp := httpResp.Header.Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'none'") {
		t.Errorf("expected Content-Security-Policy header containing default-src 'none', got '%s'", csp)
	}

	// 2. Send message with ONLY an image file attachment (.png)
	imgBuf := &bytes.Buffer{}
	imgWriter := multipart.NewWriter(imgBuf)

	imgPart, err := imgWriter.CreateFormFile("attachment", "photo.png")
	if err != nil {
		t.Fatalf("failed to create img form file: %v", err)
	}
	// fake png header bytes
	_, _ = imgPart.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))
	_ = imgWriter.Close()

	respImgMsg, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", imgWriter.FormDataContentType()).
		SetHeader("HX-Request", "true").
		SetRawBody(imgBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("post image attachment failed: %v", err)
	}
	respImgMsg.Close()

	if respImgMsg.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for image attachment message, got %d", respImgMsg.StatusCode())
	}

	// 3. Send message with Word document (.docx), Code file (.js), and SVG file (.svg)
	docxBuf := &bytes.Buffer{}
	docxWriter := multipart.NewWriter(docxBuf)
	_ = docxWriter.WriteField("content", "Here is a Word spec and script.")
	docxPart, _ := docxWriter.CreateFormFile("attachment", "spec.docx")
	_, _ = docxPart.Write([]byte("fake docx content"))
	_ = docxWriter.Close()

	respDocx, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", docxWriter.FormDataContentType()).
		SetRawBody(docxBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("post docx failed: %v", err)
	}
	respDocx.Close()

	jsBuf := &bytes.Buffer{}
	jsWriter := multipart.NewWriter(jsBuf)
	jsPart, _ := jsWriter.CreateFormFile("attachment", "script.js")
	_, _ = jsPart.Write([]byte("console.log('hello');"))
	_ = jsWriter.Close()

	respJS, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", jsWriter.FormDataContentType()).
		SetRawBody(jsBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("post js file failed: %v", err)
	}
	respJS.Close()

	svgBuf := &bytes.Buffer{}
	svgWriter := multipart.NewWriter(svgBuf)
	svgPart, _ := svgWriter.CreateFormFile("attachment", "vector.svg")
	_, _ = svgPart.Write([]byte("<svg xmlns='http://www.w3.org/2000/svg'></svg>"))
	_ = svgWriter.Close()

	respSVG, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("Content-Type", svgWriter.FormDataContentType()).
		SetRawBody(svgBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("post svg file failed: %v", err)
	}
	respSVG.Close()

	// Verify messages in database
	msgs = ch.MessagesPage(1, 10)
	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages in channel, got %d", len(msgs))
	}

	// Message index 1: photo.png -> IsImageAttachment() must be true
	if !msgs[1].IsImageAttachment() || msgs[1].AttachmentName != "photo.png" {
		t.Errorf("expected photo.png to be image attachment, got %v", msgs[1])
	}
	// Message index 2: spec.docx -> IsImageAttachment() must be false
	if msgs[2].IsImageAttachment() || msgs[2].AttachmentName != "spec.docx" {
		t.Errorf("expected spec.docx NOT to be image attachment, got %v", msgs[2])
	}
	// Message index 3: script.js -> IsImageAttachment() must be false
	if msgs[3].IsImageAttachment() || msgs[3].AttachmentName != "script.js" {
		t.Errorf("expected script.js NOT to be image attachment, got %v", msgs[3])
	}
	// Message index 4: vector.svg -> IsImageAttachment() MUST BE FALSE (prevent inline SVG script execution)
	if msgs[4].IsImageAttachment() || msgs[4].AttachmentName != "vector.svg" {
		t.Errorf("expected vector.svg NOT to be image attachment, got %v", msgs[4])
	}

	// 4. Fetch HTML page for messages and verify rendered elements
	respPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET channel messages page failed: %v", err)
	}
	htmlStr := respPage.String()
	respPage.Close()

	if respPage.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for message page render, got %d", respPage.StatusCode())
	}

	// Verify image preview element AND download overlay rendered for photo.png
	if !strings.Contains(htmlStr, "photo.png") || !strings.Contains(htmlStr, "<img src=\"/uploads/") {
		t.Errorf("expected HTML to render image tag for photo.png, got:\n%s", htmlStr)
	}
	if !strings.Contains(htmlStr, "download=\"photo.png\"") {
		t.Errorf("expected HTML to render image download button for photo.png")
	}

	// Verify document download cards rendered for notes.txt, spec.docx, script.js, vector.svg
	if !strings.Contains(htmlStr, "download=\"notes.txt\"") {
		t.Errorf("expected HTML to render download card for notes.txt")
	}
	if !strings.Contains(htmlStr, "download=\"spec.docx\"") {
		t.Errorf("expected HTML to render download card for spec.docx")
	}
	if !strings.Contains(htmlStr, "download=\"script.js\"") {
		t.Errorf("expected HTML to render download card for script.js")
	}
	if !strings.Contains(htmlStr, "download=\"vector.svg\"") {
		t.Errorf("expected HTML to render download card for vector.svg")
	}
}
