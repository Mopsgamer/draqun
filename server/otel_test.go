package internal

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestOtelChannelAccessAndMetrics(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create admin user (first user gets Draqun group admin)
	cookieAdmin, adminUser := createTestUserCookie(t, ts, "OTel Admin", "oteladmin", "oteladmin@example.com")
	draqunGroup, err := model.GetOrCreateDraqunGroup(adminUser.Id)
	if err != nil || draqunGroup.IsEmpty() {
		t.Fatalf("failed to get or create draqun group: %v", err)
	}

	channels, err := model.GetGroupChannels(draqunGroup.Id)
	if err != nil {
		t.Fatalf("failed to get group channels: %v", err)
	}

	var otelCh model.Channel
	for _, ch := range channels {
		if ch.IsOtel() {
			otelCh = ch
			break
		}
	}
	if otelCh.IsEmpty() {
		t.Fatalf("expected otel channel to exist in Draqun group")
	}

	// 1. Regular user registration
	cookieRegular, regUser := createTestUserCookie(t, ts, "OTel Regular", "otelregular", "otelregular@example.com")
	_, _ = ts.Client.R().SetHeader("Cookie", cookieRegular).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))

	// Non-admin accessing otel channel should be redirected to group page
	respReg, err := ts.Client.R().SetHeader("Cookie", cookieRegular).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, otelCh.Id))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respReg.StatusCode() != http.StatusFound && respReg.StatusCode() != http.StatusOK && respReg.StatusCode() != http.StatusSeeOther {
		t.Fatalf("expected status redirect or ok, got %d", respReg.StatusCode())
	}
	// Check content does not contain OpenTelemetry Monitor title for non-admin
	if respReg.StatusCode() == http.StatusOK && string(respReg.Body()) != "" {
		if containsString(string(respReg.Body()), "OpenTelemetry Monitor") {
			t.Fatalf("non-admin should not see OpenTelemetry Monitor dashboard")
		}
	}

	// Record sample span
	model.RecordOtelSpan("TEST SPAN", "INTERNAL", 200, 15*time.Millisecond, map[string]string{"test": "true"})

	// Global Admin accessing otel channel should get OTel dashboard with 200 OK
	respAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, otelCh.Id))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respAdmin.StatusCode() != http.StatusOK {
		t.Fatalf("admin expected 200 OK accessing otel channel, got %d", respAdmin.StatusCode())
	}

	bodyAdmin := string(respAdmin.Body())
	if !containsString(bodyAdmin, "OpenTelemetry Monitor") {
		t.Errorf("admin response should contain OpenTelemetry Monitor dashboard")
	}

	_ = regUser
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || (len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()))
}
