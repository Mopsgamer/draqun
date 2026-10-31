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

	otelCh := model.GetOtelVirtualChannel()
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
		if containsString(string(respReg.Body()), "OpenTelemetry Trace Spans") {
			t.Fatalf("non-admin should not see OpenTelemetry Trace Spans dashboard")
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
	if !containsString(bodyAdmin, "OpenTelemetry Trace Spans") {
		t.Errorf("admin response should contain OpenTelemetry Trace Spans dashboard")
	}

	_ = regUser
}

func TestOtelUniqueErrorCollection(t *testing.T) {
	// Record spans with various errors
	model.RecordOtelSpan("GET /endpoint-1", "HTTP SERVER", 500, 10*time.Millisecond, map[string]string{
		"error": "database connection refused",
	})
	// Duplicate request with error
	model.RecordOtelSpan("GET /endpoint-1", "HTTP SERVER", 500, 12*time.Millisecond, map[string]string{
		"error": "database connection refused",
	})
	// Second unique request with error
	model.RecordOtelSpan("GET /endpoint-2", "HTTP SERVER", 404, 5*time.Millisecond, map[string]string{
		"error": "resource not found",
	})
	// Duplicate of second request
	model.RecordOtelSpan("GET /endpoint-2", "HTTP SERVER", 404, 8*time.Millisecond, map[string]string{
		"error": "resource not found",
	})
	// Error from template render
	model.RecordOtelSpan("TEMPLATE RENDER: partials/home", "TEMPLATE RENDER", 500, 2*time.Millisecond, map[string]string{
		"template.error": "template syntax error",
	})
	// OK span should not add error request
	model.RecordOtelSpan("GET /endpoint-ok", "HTTP SERVER", 200, 5*time.Millisecond, nil)

	stats := model.GetOtelStats()

	// Verify unique error requests list contains the unique requests without duplicates
	foundEndpoint1 := false
	foundEndpoint2 := false
	foundTemplate := false
	endpoint1Count := 0

	for _, span := range stats.UniqueErrors {
		if span.Name == "GET /endpoint-1" {
			foundEndpoint1 = true
			endpoint1Count++
		}
		if span.Name == "GET /endpoint-2" {
			foundEndpoint2 = true
		}
		if span.Name == "TEMPLATE RENDER: partials/home" {
			foundTemplate = true
		}
	}

	if !foundEndpoint1 || endpoint1Count != 1 {
		t.Errorf("expected 'GET /endpoint-1' error request to be recorded exactly once, count=%d", endpoint1Count)
	}
	if !foundEndpoint2 {
		t.Errorf("expected 'GET /endpoint-2' in unique error requests")
	}
	if !foundTemplate {
		t.Errorf("expected 'TEMPLATE RENDER: partials/home' in unique error requests")
	}

	// Test forgetting/clearing errors
	model.ClearOtelErrors()
	clearedStats := model.GetOtelStats()
	if len(clearedStats.UniqueErrors) != 0 {
		t.Errorf("expected 0 unique errors after clear, got %d", len(clearedStats.UniqueErrors))
	}
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
