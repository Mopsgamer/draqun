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

func TestServerConfigVirtualChannelAndLockdown(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create admin user (first user becomes global admin)
	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Config Admin", "cfgadmin", "cfgadmin@example.com")
	draqunGroup, err := model.GetOrCreateDraqunGroup(adminUser.Id)
	if err != nil || draqunGroup.IsEmpty() {
		t.Fatalf("failed to get or create draqun group: %v", err)
	}

	configCh := model.GetConfigVirtualChannel()
	if configCh.IsEmpty() {
		t.Fatalf("expected config channel to exist")
	}

	// 1. Regular user registration
	cookieRegular, regUser := createTestUserCookie(t, ts, "Config Regular", "cfgregular", "cfgregular@example.com")
	_, _ = ts.Client.R().SetHeader("Cookie", cookieRegular).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))

	// Non-admin accessing config channel should be redirected to group page
	respReg, err := ts.Client.R().SetHeader("Cookie", cookieRegular).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, configCh.Id))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respReg.StatusCode() == http.StatusOK && containsString(string(respReg.Body()), "Server Configuration") {
		t.Fatalf("non-admin should not see Server Configuration dashboard")
	}

	// Global Admin accessing config channel should see dashboard with 200 OK
	respAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, configCh.Id))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respAdmin.StatusCode() != http.StatusOK {
		t.Fatalf("admin expected 200 OK accessing config channel, got %d", respAdmin.StatusCode())
	}
	bodyAdmin := string(respAdmin.Body())
	if !containsString(bodyAdmin, "Server Configuration") {
		t.Errorf("admin response should contain Server Configuration dashboard")
	}

	// 2. Global Admin updates Server Configuration via POST /groups/:group_id/config (max_file_size = 1 MB)
	updateResp, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"max_message_size":      "500",
			"user_token_expiration": "240",
			"max_file_size":         "1",
			"notice_message":        "**Maintenance Notice:** System will restart soon.",
			"lockdown":              "true",
		}).
		Post(fmt.Sprintf("/groups/%d/config", draqunGroup.Id))
	if err != nil {
		t.Fatalf("failed to update config: %v", err)
	}
	if updateResp.StatusCode() != http.StatusOK && updateResp.StatusCode() != http.StatusFound && updateResp.StatusCode() != http.StatusSeeOther {
		t.Fatalf("expected 200 OK, 302 or 303 Redirect for config update, got %d", updateResp.StatusCode())
	}

	// Verify updated settings in DB model
	cfg := model.GetServerConfig()
	if cfg.MaxMessageSize != 500 {
		t.Errorf("expected MaxMessageSize 500, got %d", cfg.MaxMessageSize)
	}
	if cfg.UserTokenExpiration != 240 {
		t.Errorf("expected UserTokenExpiration 240, got %d", cfg.UserTokenExpiration)
	}
	if cfg.MaxFileSize != 1 {
		t.Errorf("expected MaxFileSize 1 MB, got %d", cfg.MaxFileSize)
	}
	if cfg.MaxFileSizeBytes() != 1024*1024 {
		t.Errorf("expected MaxFileSizeBytes %d, got %d", 1024*1024, cfg.MaxFileSizeBytes())
	}
	if !cfg.Lockdown {
		t.Errorf("expected Lockdown true, got false")
	}
	if cfg.NoticeMessage != "**Maintenance Notice:** System will restart soon." {
		t.Errorf("unexpected notice message: %s", cfg.NoticeMessage)
	}

	// 3. Verify Notice Header banner rendered on page and red (lockdown = true)
	respPage, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get("/chat")
	if err != nil {
		t.Fatalf("failed to fetch chat page: %v", err)
	}
	bodyPage := string(respPage.Body())
	if !containsString(bodyPage, "notice-header") {
		t.Errorf("notice-header missing from page body")
	}
	if !containsString(bodyPage, "bg-red-950") {
		t.Errorf("notice-header should be red when lockdown is enabled")
	}

	// 4. Verify Lockdown blocks mutating state actions for non-admin
	respLockdown, err := ts.Client.R().
		SetHeader("Cookie", cookieRegular).
		SetFormDataWithMap(map[string]string{
			"content": "Hello in lockdown",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/1/messages/create", draqunGroup.Id))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respLockdown.StatusCode() != http.StatusForbidden && !strings.Contains(string(respLockdown.Body()), "lockdown") {
		t.Errorf("expected lockdown block for regular user, got status %d body %s", respLockdown.StatusCode(), string(respLockdown.Body()))
	}

	// 5. Test File Upload Max Size limit in MB
	// Create a multipart upload of 2 MB exceeding max_file_size (1 MB)
	largeData := bytes.Repeat([]byte("A"), 2*1024*1024)
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)
	part, _ := writer.CreateFormFile("attachments", "large.txt")
	part.Write(largeData)
	writer.WriteField("content", "message with 2MB file")
	writer.Close()

	// Disable lockdown temporarily to test file size check specifically
	cfg.Lockdown = false
	_ = cfg.Update()

	respFile, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetRawBody(bodyBuf.Bytes()).
		Post(fmt.Sprintf("/groups/%d/channels/1/messages/create", draqunGroup.Id))
	if err != nil {
		t.Fatalf("file upload request failed: %v", err)
	}
	bodyFile := string(respFile.Body())
	if !containsString(bodyFile, "exceeds the maximum allowed limit") && !containsString(bodyFile, "file too large") && !strings.Contains(bodyFile, "limit") {
		t.Errorf("expected file size limit rejection for 2MB file with 1MB limit, got response: %s", bodyFile)
	}

	// 6. Test Teal Notice Header when lockdown = false
	respTeal, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get("/chat")
	if err != nil {
		t.Fatalf("failed to fetch chat page: %v", err)
	}
	bodyTeal := string(respTeal.Body())
	if !containsString(bodyTeal, "bg-teal-950") {
		t.Errorf("notice-header should be teal when lockdown is disabled")
	}

	// 7. Test Lockdown enabled with empty notice_message -> shows default red notice
	cfg.Lockdown = true
	cfg.NoticeMessage = ""
	_ = cfg.Update()

	respEmptyLockdown, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get("/chat")
	if err != nil {
		t.Fatalf("failed to fetch chat page: %v", err)
	}
	bodyEmptyLockdown := string(respEmptyLockdown.Body())
	if !containsString(bodyEmptyLockdown, "notice-header") || !containsString(bodyEmptyLockdown, "bg-red-950") || !containsString(bodyEmptyLockdown, "Server Lockdown") {
		t.Errorf("expected red notice header with default lockdown message when notice_message is empty during lockdown")
	}

	// 8. Test Lockdown disabled with empty notice_message -> notice is not rendered
	cfg.Lockdown = false
	cfg.NoticeMessage = ""
	_ = cfg.Update()

	respNoNotice, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get("/chat")
	if err != nil {
		t.Fatalf("failed to fetch chat page: %v", err)
	}
	bodyNoNotice := string(respNoNotice.Body())
	if containsString(bodyNoNotice, "notice-header") {
		t.Errorf("notice-header should NOT be rendered when lockdown is disabled and notice_message is empty")
	}

	_ = regUser
}
