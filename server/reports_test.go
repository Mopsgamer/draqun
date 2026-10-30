package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestReportingFlow(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register test users and group
	cookieReporter, reporter := createTestUserCookie(t, ts, "Reporter User", "reporter_usr", "reporter@example.com")
	_, reportedUser := createTestUserCookie(t, ts, "Reported User", "reported_usr", "reported@example.com")
	group := createTestGroupDirectly(t, int64(reportedUser.Id), "Reported Group", "reported_group")

	// Create a test channel in group
	channel := model.NewChannel("general", "General", "General channel", model.ChannelTypeText)
	if err := channel.Insert(); err != nil {
		t.Fatalf("failed to insert test channel: %v", err)
	}
	_ = model.AddGroupChannel(group.Id, channel.Id, 1)

	msg := model.NewMessageFilled(channel.Id, group.Id, reportedUser.Id, "Inappropriate message content")
	if err := msg.Insert(); err != nil {
		t.Fatalf("failed to insert test message: %v", err)
	}

	// 1. Test Report Group
	t.Run("Report Group", func(t *testing.T) {
		resp, err := ts.Client.R().
			SetHeader("Cookie", cookieReporter).
			SetFormDataWithMap(map[string]string{
				"reason": "Group contains spam content",
			}).
			Post(fmt.Sprintf("/reports/group/%d", group.Id))

		if err != nil {
			t.Fatalf("failed to send report group request: %v", err)
		}
		resp.Close()
		if resp.StatusCode() != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode())
		}
	})

	// 2. Test Report User
	t.Run("Report User", func(t *testing.T) {
		resp, err := ts.Client.R().
			SetHeader("Cookie", cookieReporter).
			SetFormDataWithMap(map[string]string{
				"reason": "Harassment and abuse",
			}).
			Post(fmt.Sprintf("/reports/user/%d", reportedUser.Id))

		if err != nil {
			t.Fatalf("failed to send report user request: %v", err)
		}
		resp.Close()
		if resp.StatusCode() != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode())
		}
	})

	// 3. Test Report Message
	t.Run("Report Message", func(t *testing.T) {
		resp, err := ts.Client.R().
			SetHeader("Cookie", cookieReporter).
			SetFormDataWithMap(map[string]string{
				"reason": "Offensive language in message",
			}).
			Post(fmt.Sprintf("/reports/message/%d", msg.Id))

		if err != nil {
			t.Fatalf("failed to send report message request: %v", err)
		}
		resp.Close()
		if resp.StatusCode() != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode())
		}
	})

	// Verify that tickets were created in Draqun group's reports channel
	draqunGroup, err := model.GetOrCreateDraqunGroup(reporter.Id)
	if err != nil {
		t.Fatalf("failed to get Draqun group: %v", err)
	}

	reportsChannel, alert := model.GetOrCreateReportsChannel(draqunGroup.Id)
	if alert != nil || reportsChannel.IsEmpty() {
		t.Fatalf("failed to get reports channel in Draqun group")
	}

	subchannels, err := model.GetSubchannels(reportsChannel.Id)
	if err != nil {
		t.Fatalf("failed to get subchannels for reports channel: %v", err)
	}

	if len(subchannels) < 3 {
		t.Fatalf("expected at least 3 report ticket subchannels, got %d", len(subchannels))
	}
}
