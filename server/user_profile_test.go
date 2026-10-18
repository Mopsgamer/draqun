package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestUserProfileModal(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "adminuser", "admin@example.com")
	_, targetUser := createTestUserCookie(t, ts, "Target User", "targetuser", "target@example.com")

	// Ensure admin has Draqun Admin status by creating/getting draqun group with adminUser as owner
	_, _ = model.GetOrCreateDraqunGroup(adminUser.Id)

	// GET target user's profile modal
	resp, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/users/%d/profile", targetUser.Id))
	if err != nil {
		t.Fatalf("GET /users/:id/profile failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(bodyStr, "Target User") {
		t.Errorf("expected target user moniker in profile modal, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "@targetuser") {
		t.Errorf("expected target user name in profile modal, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Account Created") {
		t.Errorf("expected account created date in profile modal, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "copyToClipboard") {
		t.Errorf("expected copy to clipboard function in profile modal, got: %s", bodyStr)
	}
}

func TestUserResetAvatarAndMoniker(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "adminuser2", "admin2@example.com")
	_, targetUser := createTestUserCookie(t, ts, "Bad Moniker", "badmoniker", "bad@example.com")

	_, _ = model.GetOrCreateDraqunGroup(adminUser.Id)

	// Reset Avatar
	respAvatar, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		Post(fmt.Sprintf("/admin/users/%d/reset-avatar", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/reset-avatar failed: %v", err)
	}
	defer respAvatar.Close()

	if respAvatar.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respAvatar.StatusCode())
	}

	updatedUser, _ := model.NewUserFromId(targetUser.Id)
	if updatedUser.Avatar != "" {
		t.Errorf("expected avatar to be reset to empty, got %q", updatedUser.Avatar)
	}

	// Reset Moniker
	respMoniker, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		Post(fmt.Sprintf("/admin/users/%d/reset-moniker", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/reset-moniker failed: %v", err)
	}
	defer respMoniker.Close()

	if respMoniker.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respMoniker.StatusCode())
	}

	updatedUser, _ = model.NewUserFromId(targetUser.Id)
	expectedMoniker := fmt.Sprintf("User_%d", targetUser.Id)
	if string(updatedUser.Moniker) != expectedMoniker {
		t.Errorf("expected moniker to be reset to %s, got %s", expectedMoniker, updatedUser.Moniker)
	}

	// Reset Name (@username)
	respName, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		Post(fmt.Sprintf("/admin/users/%d/reset-name", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/reset-name failed: %v", err)
	}
	defer respName.Close()

	if respName.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respName.StatusCode())
	}

	updatedUser, _ = model.NewUserFromId(targetUser.Id)
	expectedName := fmt.Sprintf("user_%d", targetUser.Id)
	if string(updatedUser.Name) != expectedName {
		t.Errorf("expected name to be reset to %s, got %s", expectedName, updatedUser.Name)
	}

	// Verify logs
	logs := model.GetUserLogs(targetUser.Id, 10)
	if len(logs) < 3 {
		t.Errorf("expected at least 3 audit logs, got %d", len(logs))
	}
}

func TestUserAccountSuspension(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "adminuser3", "admin3@example.com")
	_, targetUser := createTestUserCookie(t, ts, "Spammer User", "spammer", "spammer@example.com")

	group, _ := model.GetOrCreateDraqunGroup(adminUser.Id)

	// Suspend Account
	respSuspend, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"reason": "Spamming community channels",
		}).
		Post(fmt.Sprintf("/admin/users/%d/suspend", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/suspend failed: %v", err)
	}
	defer respSuspend.Close()

	if respSuspend.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respSuspend.StatusCode())
	}

	updatedUser, _ := model.NewUserFromId(targetUser.Id)
	if !updatedUser.IsSuspended {
		t.Errorf("expected IsSuspended to be true")
	}
	if updatedUser.SuspensionReason != "Spamming community channels" {
		t.Errorf("expected SuspensionReason to match, got %q", updatedUser.SuspensionReason)
	}

	if !group.IsBanned(targetUser.Id) {
		t.Errorf("expected suspended user to be banned across groups")
	}

	// Unsuspend
	respUnsuspend, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		Post(fmt.Sprintf("/admin/users/%d/unsuspend", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/unsuspend failed: %v", err)
	}
	defer respUnsuspend.Close()

	if respUnsuspend.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respUnsuspend.StatusCode())
	}

	updatedUser, _ = model.NewUserFromId(targetUser.Id)
	if updatedUser.IsSuspended {
		t.Errorf("expected IsSuspended to be false after unsuspend")
	}
}

func TestUserAccountLockdown(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "adminuser4", "admin4@example.com")
	cookieTarget, targetUser := createTestUserCookie(t, ts, "Hacked User", "hacked", "hacked@example.com")

	_, _ = model.GetOrCreateDraqunGroup(adminUser.Id)

	// Lockdown Account
	respLock, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		Post(fmt.Sprintf("/admin/users/%d/lockdown", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/lockdown failed: %v", err)
	}
	defer respLock.Close()

	if respLock.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respLock.StatusCode())
	}

	updatedUser, _ := model.NewUserFromId(targetUser.Id)
	if !updatedUser.IsLocked {
		t.Errorf("expected IsLocked to be true")
	}

	// Locked user attempts guarded API request -> auth middleware should reject with AlertAccountLocked (403 Forbidden)
	respAttempt, err := ts.Client.R().
		SetHeader("Cookie", cookieTarget).
		Post("/groups/create")
	if err != nil {
		t.Fatalf("POST /groups/create failed: %v", err)
	}
	defer respAttempt.Close()

	if respAttempt.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for locked user, got %d", respAttempt.StatusCode())
	}

	// Locked user attempting login should get AlertAccountLocked alert
	respLogin, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"email":    "hacked@example.com",
			"password": "password123",
		}).
		Post("/account/login")
	if err != nil {
		t.Fatalf("POST /account/login for locked user failed: %v", err)
	}
	defer respLogin.Close()

	if respLogin.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when logging in as locked user, got %d", respLogin.StatusCode())
	}

	// Unlock Account
	respUnlock, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		Post(fmt.Sprintf("/admin/users/%d/unlock", targetUser.Id))
	if err != nil {
		t.Fatalf("POST /admin/users/:id/unlock failed: %v", err)
	}
	defer respUnlock.Close()

	if respUnlock.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respUnlock.StatusCode())
	}

	updatedUser, _ = model.NewUserFromId(targetUser.Id)
	if updatedUser.IsLocked {
		t.Errorf("expected IsLocked to be false after unlock")
	}
}

func TestIPBlocking(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "adminuser5", "admin5@example.com")
	_, _ = model.GetOrCreateDraqunGroup(adminUser.Id)

	// Attempt to block 127.0.0.1 -> should fail / return error
	respLocalhost, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"ip":     "127.0.0.1",
			"reason": "Test block localhost",
		}).
		Post("/admin/ips/block")
	if err != nil {
		t.Fatalf("POST /admin/ips/block failed: %v", err)
	}
	defer respLocalhost.Close()

	if respLocalhost.StatusCode() == http.StatusOK {
		t.Errorf("expected failure when blocking localhost 127.0.0.1")
	}
	if model.IsIPBlocked("127.0.0.1") {
		t.Errorf("localhost 127.0.0.1 must never be blocked")
	}

	// Block a valid remote IP 203.0.113.5
	testIP := "203.0.113.5"
	respBlock, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"ip":     testIP,
			"reason": "Malicious activity",
		}).
		Post("/admin/ips/block")
	if err != nil {
		t.Fatalf("POST /admin/ips/block failed: %v", err)
	}
	defer respBlock.Close()

	if respBlock.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respBlock.StatusCode())
	}

	if !model.IsIPBlocked(testIP) {
		t.Errorf("expected IP %s to be blocked", testIP)
	}

	// Unblock IP
	respUnblock, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"ip": testIP,
		}).
		Post("/admin/ips/unblock")
	if err != nil {
		t.Fatalf("POST /admin/ips/unblock failed: %v", err)
	}
	defer respUnblock.Close()

	if respUnblock.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respUnblock.StatusCode())
	}

	if model.IsIPBlocked(testIP) {
		t.Errorf("expected IP %s to be unblocked", testIP)
	}
}

func TestBannedUserRedirectAndJoinPageReason(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "adminuser6", "admin6@example.com")
	cookieTarget, targetUser := createTestUserCookie(t, ts, "Banned User", "banneduser", "banned@example.com")

	group := createTestGroupDirectly(t, int64(adminUser.Id), "Secret Group", "secret_group")

	// Ban target user from group
	respBan, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"description":   "Rule breaking in chat",
			"duration_days": "30",
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/ban", group.Id, targetUser.Id))
	if err != nil {
		t.Fatalf("POST /groups/:id/members/:id/ban failed: %v", err)
	}
	defer respBan.Close()

	// Banned user trying to open /chat/groups/:id -> should redirect to /groups/join/secret_group
	respChat, err := ts.Client.R().
		SetHeader("Cookie", cookieTarget).
		Get(fmt.Sprintf("/chat/groups/%d", group.Id))
	if err != nil {
		t.Fatalf("GET /chat/groups/:id failed: %v", err)
	}
	defer respChat.Close()

	if respChat.StatusCode() != http.StatusSeeOther && respChat.StatusCode() != http.StatusFound {
		t.Fatalf("expected 302/303 redirect for banned user, got %d", respChat.StatusCode())
	}
	location := respChat.Header("Location")
	if !strings.Contains(location, "/chat/groups/join/") {
		t.Errorf("expected redirect location to contain /chat/groups/join/, got %s", location)
	}

	// GET /chat/groups/join/secret_group as banned user -> should contain ban reason and expiration
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieTarget).
		Get(fmt.Sprintf("/chat/groups/join/%s", group.Name))
	if err != nil {
		t.Fatalf("GET /groups/join/:name failed: %v", err)
	}
	defer respJoin.Close()

	if respJoin.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for join page, got %d", respJoin.StatusCode())
	}

	bodyStr := respJoin.String()
	if !strings.Contains(bodyStr, "You are banned from this group") {
		t.Errorf("expected ban warning banner on join page, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Rule breaking in chat") {
		t.Errorf("expected ban reason 'Rule breaking in chat', got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Duration:") || !strings.Contains(bodyStr, "30 days") {
		t.Errorf("expected duration '30 days' on join page, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Ban Expiration:") {
		t.Errorf("expected ban expiration date on join page, got: %s", bodyStr)
	}

	// Verify ban notification for target user
	notifs, errN := model.GetUserNotifications(targetUser.Id, 10)
	if errN != nil || len(notifs) == 0 {
		t.Fatalf("expected ban notification for target user, got err: %v, count: %d", errN, len(notifs))
	}
	foundBanNotif := false
	for _, n := range notifs {
		if n.Type == "banned" && strings.Contains(n.Content, "Rule breaking in chat") && strings.Contains(n.Content, "30 days") {
			foundBanNotif = true
			break
		}
	}
	if !foundBanNotif {
		t.Errorf("expected notification containing reason 'Rule breaking in chat' and duration '30 days', got notifs: %+v", notifs)
	}

	// Test banning without a reason -> should NOT display Reason: on join page or in notification
	_, noReasonUser := createTestUserCookie(t, ts, "NoReason User", "noreasonuser", "noreason@example.com")
	respBanNoReason, err := ts.Client.R().
		SetHeader("Cookie", cookieAdmin).
		SetFormDataWithMap(map[string]string{
			"duration_days": "10",
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/ban", group.Id, noReasonUser.Id))
	if err != nil {
		t.Fatalf("POST /groups/:id/members/:id/ban without reason failed: %v", err)
	}
	defer respBanNoReason.Close()

	respJoinNoReason, err := ts.Client.R().
		SetHeader("Cookie", createTestUserCookieHeader(t, ts, noReasonUser)).
		Get(fmt.Sprintf("/chat/groups/join/%s", group.Name))
	if err != nil {
		t.Fatalf("GET /groups/join/:name failed: %v", err)
	}
	defer respJoinNoReason.Close()

	bodyNoReason := respJoinNoReason.String()
	if strings.Contains(bodyNoReason, "Reason:") {
		t.Errorf("expected join page NOT to contain 'Reason:' when no reason was provided, got: %s", bodyNoReason)
	}
	if !strings.Contains(bodyNoReason, "Duration:") || !strings.Contains(bodyNoReason, "10 days") {
		t.Errorf("expected join page to contain duration '10 days', got: %s", bodyNoReason)
	}

	notifsNoReason, _ := model.GetUserNotifications(noReasonUser.Id, 10)
	if len(notifsNoReason) == 0 {
		t.Fatalf("expected notification for noReasonUser")
	}
	if strings.Contains(notifsNoReason[0].Content, "Reason:") {
		t.Errorf("expected notification NOT to contain 'Reason:', got: %s", notifsNoReason[0].Content)
	}
	if !strings.Contains(notifsNoReason[0].Content, "Duration: 10 days") {
		t.Errorf("expected notification to contain 'Duration: 10 days', got: %s", notifsNoReason[0].Content)
	}
}

func createTestUserCookieHeader(t *testing.T, ts *TestServer, user model.User) string {
	token, _ := user.GenerateToken()
	return fmt.Sprintf("Authorization=%s", token)
}
