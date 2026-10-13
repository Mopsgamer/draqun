package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestNotificationAndMentions(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// 1. Setup users and group
	cookie1, user1 := createTestUserCookie(t, ts, "Notif User One", "notif_user1", "notif_user1@example.com")
	cookie2, user2 := createTestUserCookie(t, ts, "Notif User Two", "notif_user2", "notif_user2@example.com")

	group := createTestGroupDirectly(t, int64(user1.Id), "Notif Group", "notif_group")

	// Add user2 to group
	member2 := model.NewMember(group.Id, user2.Id, "")
	_ = member2.Insert()

	channel := model.NewChannel("gen_notif", "General Notif", "desc", model.ChannelTypeText)
	_ = channel.Insert()
	_ = model.AddGroupChannel(group.Id, channel.Id, 0)

	// 2. Test mention extraction
	mentioned, err := model.ExtractMentionedUsers(group.Id, "Hello @notif_user2 welcome!")
	if err != nil {
		t.Fatalf("failed to extract mentioned users: %v", err)
	}
	if len(mentioned) != 1 || mentioned[0].Id != user2.Id {
		t.Fatalf("expected user2 to be mentioned, got: %v", mentioned)
	}

	// 3. Process message mentions
	msg := model.NewMessageFilled(channel.Id, group.Id, user1.Id, "Hey @notif_user2 check this out")
	_ = msg.Insert()
	model.ProcessMessageMentions(app, msg)

	// 4. Verify user2 received notification
	unread, err := model.GetUnreadNotificationCount(user2.Id)
	if err != nil {
		t.Fatalf("failed to get unread count: %v", err)
	}
	if unread != 1 {
		t.Fatalf("expected 1 unread notification for user2, got: %d", unread)
	}

	notifs, err := model.GetUserNotifications(user2.Id, 10)
	if err != nil || len(notifs) != 1 {
		t.Fatalf("expected 1 notification in list, got %d (err: %v)", len(notifs), err)
	}
	if !strings.Contains(notifs[0].Title, string(user1.Moniker)) {
		t.Fatalf("expected notification title to contain author moniker, got: %s", notifs[0].Title)
	}

	// 5. Test GET /notifications partial route (as user2)
	respGet, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Get("/notifications")
	if err != nil || respGet.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from GET /notifications, got status: %v", respGet.StatusCode())
	}

	// Test GET /notifications full page route
	respGetPage, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		Get("/notifications")
	if err != nil || respGetPage.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from GET /notifications full page, got status: %v", respGetPage.StatusCode())
	}

	// 6. Test POST /notifications/:id/read route
	reqReadUrl := fmt.Sprintf("/notifications/%d/read", notifs[0].Id)
	respRead, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Post(reqReadUrl)
	if err != nil || respRead.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from POST /notifications/:id/read, got status: %v", respRead.StatusCode())
	}

	unreadAfter, _ := model.GetUnreadNotificationCount(user2.Id)
	if unreadAfter != 0 {
		t.Fatalf("expected 0 unread notifications after marking read, got: %d", unreadAfter)
	}

	// 7. Test Ticket Assignment Notification
	ticketCh := model.NewChannel("issue_1", "Bug Report 1", "bug", model.ChannelTypeTicket)
	_ = ticketCh.Insert()
	_ = model.AddGroupChannel(group.Id, ticketCh.Id, 1)

	respAssign, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetFormDataWithMap(map[string]string{
			"user_id": fmt.Sprintf("%d", user2.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/assignees", group.Id, ticketCh.Id))
	if err != nil || (respAssign.StatusCode() != http.StatusOK && respAssign.StatusCode() != http.StatusSeeOther) {
		t.Fatalf("expected success from ticket assign, got: %d", respAssign.StatusCode())
	}

	notifsUser2, _ := model.GetUserNotifications(user2.Id, 10)
	if len(notifsUser2) < 2 || notifsUser2[0].Type != "ticket_assigned" {
		t.Fatalf("expected ticket_assigned notification for user2, got: %v", notifsUser2)
	}
	if notifsUser2[0].ChannelId != ticketCh.Id {
		t.Fatalf("expected notification channelId to be %d, got %d", ticketCh.Id, notifsUser2[0].ChannelId)
	}

	// 8. Test DELETE /notifications/:id (Single deletion)
	respDelSingle, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Delete(fmt.Sprintf("/notifications/%d", notifsUser2[0].Id))
	if err != nil || respDelSingle.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from single notification delete, got: %d", respDelSingle.StatusCode())
	}

	notifsAfterDel, _ := model.GetUserNotifications(user2.Id, 10)
	if len(notifsAfterDel) != 1 {
		t.Fatalf("expected 1 notification remaining after single delete, got: %d", len(notifsAfterDel))
	}

	// 9. Test DELETE /notifications/all (Delete all notifications)
	respDelAll, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Delete("/notifications/all")
	if err != nil || respDelAll.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from delete all notifications, got: %d", respDelAll.StatusCode())
	}

	notifsFinal, _ := model.GetUserNotifications(user2.Id, 10)
	if len(notifsFinal) != 0 {
		t.Fatalf("expected 0 notifications after delete all, got: %d", len(notifsFinal))
	}

	_ = cookie1
}
