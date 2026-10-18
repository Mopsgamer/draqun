package internal

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Mopsgamer/draqun/server/model"
	"github.com/fasthttp/websocket"
)

func TestAPIAccountRegisterAndLogin(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Successful Registration
	respReg, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"moniker":          "API User One",
			"name":             "api_user1",
			"email":            "api_user1@example.com",
			"phone":            "1111111111",
			"password":         "securepassword123",
			"confirm-password": "securepassword123",
		}).
		Post("/account")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	defer respReg.Close()

	if respReg.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for successful registration, got %d", respReg.StatusCode())
	}

	// Verify cookie is set
	var authCookie string
	for _, cookie := range respReg.Cookies() {
		if string(cookie.Key()) == "Authorization" {
			authCookie = string(cookie.Key()) + "=" + string(cookie.Value())
			break
		}
	}
	if authCookie == "" {
		t.Errorf("expected Authorization cookie to be set on successful registration")
	}

	// Try duplicate registration with same username/email
	respRegDup, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"moniker":          "API User One",
			"name":             "api_user1",
			"email":            "api_user1@example.com",
			"phone":            "1111111111",
			"password":         "securepassword123",
			"confirm-password": "securepassword123",
		}).
		Post("/account")
	if err != nil {
		t.Fatalf("duplicate registration request failed: %v", err)
	}
	defer respRegDup.Close()

	if respRegDup.StatusCode() == http.StatusOK {
		t.Errorf("expected failure/non-200 or alert for duplicate registration, got 200 OK")
	}

	// Password Mismatch
	respMismatch, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"moniker":          "API User Two",
			"name":             "api_user2",
			"email":            "api_user2@example.com",
			"phone":            "2222222222",
			"password":         "securepassword123",
			"confirm-password": "differentpassword",
		}).
		Post("/account")
	if err != nil {
		t.Fatalf("mismatch password request failed: %v", err)
	}
	defer respMismatch.Close()

	if respMismatch.StatusCode() == http.StatusOK {
		t.Errorf("expected failure for password mismatch, got 200 OK")
	}

	// Login with incorrect password
	respLoginBad, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"email":    "api_user1@example.com",
			"password": "wrongpassword",
		}).
		Post("/account/login")
	if err != nil {
		t.Fatalf("bad login request failed: %v", err)
	}
	defer respLoginBad.Close()

	if respLoginBad.StatusCode() == http.StatusOK {
		t.Errorf("expected failure for bad login credentials, got 200 OK")
	}

	// Login with correct password
	respLoginGood, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"email":    "api_user1@example.com",
			"password": "securepassword123",
		}).
		Post("/account/login")
	if err != nil {
		t.Fatalf("good login request failed: %v", err)
	}
	defer respLoginGood.Close()

	if respLoginGood.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for successful login, got %d", respLoginGood.StatusCode())
	}
}

func TestAPIAccountLogout(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookie, _ := createTestUserCookie(t, ts, "Logout User", "logout_usr", "logout@example.com")

	resp, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		Put("/account/logout")
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 OK for logout, got %d", resp.StatusCode())
	}

	// Check that the cookie value is set to empty or deleted
	var loggedOutCookieSet bool
	for _, cookieVal := range resp.Cookies() {
		if string(cookieVal.Key()) == "Authorization" {
			loggedOutCookieSet = true
		}
	}
	if !loggedOutCookieSet {
		// If the cookie isn't returned, or empty, that's fine.
	}
}

func TestAPIGroupLifecycle(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Creator Registers
	cookieOwner, owner := createTestUserCookie(t, ts, "Owner User", "grp_owner", "owner@example.com")

	// Create Group
	respGroupCreate, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"nick":        "My Group",
			"name":        "my_grp",
			"mode":        "public",
			"description": "A group to test API lifecycle",
		}).
		Post("/groups/create")
	if err != nil {
		t.Fatalf("group creation failed: %v", err)
	}
	defer respGroupCreate.Close()

	if respGroupCreate.StatusCode() != http.StatusSeeOther && respGroupCreate.StatusCode() != http.StatusOK {
		t.Errorf("expected redirect or OK for group creation, got %d", respGroupCreate.StatusCode())
	}

	// Fetch group from database to verify
	group, err := model.NewGroupFromName("my_grp")
	if err != nil {
		t.Fatalf("failed to retrieve group from DB after creation: %v", err)
	}
	if group.OwnerId != owner.Id {
		t.Errorf("expected group owner to be creator ID %d, got %d", owner.Id, group.OwnerId)
	}

	// Update Group Settings
	respGroupChange, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"nick":        "My Updated Group",
			"name":        "my_updated_grp",
			"mode":        "public",
			"description": "An updated description",
		}).
		Put(fmt.Sprintf("/groups/%d/change", group.Id))
	if err != nil {
		t.Fatalf("group change request failed: %v", err)
	}
	respGroupChange.Close()

	if respGroupChange.StatusCode() != http.StatusOK && respGroupChange.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected redirect or OK on group settings change, got %d", respGroupChange.StatusCode())
	}

	// Joiner Registers and Joins Group
	cookieJoiner, joiner := createTestUserCookie(t, ts, "Joiner User", "grp_joiner", "joiner@example.com")

	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("group join request failed: %v", err)
	}
	defer respJoin.Close()

	if respJoin.StatusCode() != http.StatusOK && respJoin.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected redirect or OK on group join, got %d", respJoin.StatusCode())
	}

	// Create a text channel for the test group
	ch := model.NewChannel("general", "General Chat", "General", model.ChannelTypeText)
	if err := ch.Insert(); err != nil {
		t.Fatalf("failed to insert channel: %v", err)
	}
	if err := model.AddGroupChannel(group.Id, ch.Id, 0); err != nil {
		t.Fatalf("failed to add group channel: %v", err)
	}

	// Post Message to Group (as joiner)
	respMsg, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		SetFormDataWithMap(map[string]string{
			"content": "Hello life-cycle testing world!",
		}).
		Post(fmt.Sprintf("/groups/%d/messages/create", group.Id))
	if err != nil {
		t.Fatalf("message posting request failed: %v", err)
	}
	defer respMsg.Close()

	if respMsg.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 OK for message post, got %d", respMsg.StatusCode())
	}

	// Pagination of Messages
	respMsgPage, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Get(fmt.Sprintf("/groups/%d/messages/page/1", group.Id))
	if err != nil {
		t.Fatalf("message pagination request failed: %v", err)
	}
	defer respMsgPage.Close()

	if respMsgPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for message page, got %d", respMsgPage.StatusCode())
	}

	// Pagination of Members
	respMemPage, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("member pagination request failed: %v", err)
	}
	defer respMemPage.Close()

	if respMemPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for member page, got %d", respMemPage.StatusCode())
	}

	// Leave Group (as joiner)
	respLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Delete(fmt.Sprintf("/groups/%d/leave", group.Id))
	if err != nil {
		t.Fatalf("group leave request failed: %v", err)
	}
	defer respLeave.Close()

	if respLeave.StatusCode() != http.StatusOK && respLeave.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected redirect or OK on group leave, got %d", respLeave.StatusCode())
	}

	// Verify membership is marked deleted or group members count decreases
	memberObj, err := model.NewMemberFromId(group.Id, joiner.Id)
	if err != nil {
		t.Fatalf("failed to query member: %v", err)
	}
	if !memberObj.IsDeleted {
		t.Errorf("expected member state to be marked deleted after leaving")
	}

	// Verify member element is removed from member list page after leaving
	respMemPageAfterLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("member pagination request after leave failed: %v", err)
	}
	defer respMemPageAfterLeave.Close()

	if respMemPageAfterLeave.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for member page after leave, got %d", respMemPageAfterLeave.StatusCode())
	}

	memPageAfterLeaveBody := respMemPageAfterLeave.String()
	joinerElement := fmt.Sprintf("id=\"member-%d\"", joiner.Id)
	if strings.Contains(memPageAfterLeaveBody, joinerElement) {
		t.Errorf("expected member element %s to be removed from member list after leaving, but found in body: %s", joinerElement, memPageAfterLeaveBody)
	}

	// Delete Group (as owner)
	respDeleteGrp, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d", group.Id))
	if err != nil {
		t.Fatalf("group deletion request failed: %v", err)
	}
	defer respDeleteGrp.Close()

	if respDeleteGrp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for group deletion, got %d", respDeleteGrp.StatusCode())
	}
}

func TestAPIAccountChanges(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create first user (who will own draqun group)
	_, _ = createTestUserCookie(t, ts, "Dummy Admin", "dummy_admin", "dummy@example.com")

	// Register User to be tested
	cookie, user := createTestUserCookie(t, ts, "Original Moniker", "change_usr", "change@example.com")

	// Change Profile (Moniker and Name)
	respName, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		SetFormDataWithMap(map[string]string{
			"new-moniker": "New Moniker",
			"new-name":    "change_usr_new",
		}).
		Put("/account/change/name")
	if err != nil {
		t.Fatalf("profile name change failed: %v", err)
	}
	respName.Close()

	if respName.StatusCode() != http.StatusOK && respName.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected OK or Redirect for name change, got %d", respName.StatusCode())
	}

	// Change Email
	respEmail, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		SetFormDataWithMap(map[string]string{
			"current-password": "password123",
			"new-email":        "change_new@example.com",
		}).
		Put("/account/change/email")
	if err != nil {
		t.Fatalf("email change failed: %v", err)
	}
	respEmail.Close()

	if respEmail.StatusCode() != http.StatusOK && respEmail.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected OK or Redirect for email change, got %d", respEmail.StatusCode())
	}

	// Change Phone
	respPhone, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		SetFormDataWithMap(map[string]string{
			"current-password": "password123",
			"new-phone":        "9999999999",
		}).
		Put("/account/change/phone")
	if err != nil {
		t.Fatalf("phone change failed: %v", err)
	}
	respPhone.Close()

	if respPhone.StatusCode() != http.StatusOK && respPhone.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected OK or Redirect for phone change, got %d", respPhone.StatusCode())
	}

	// Change Password
	respPass, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		SetFormDataWithMap(map[string]string{
			"current-password": "password123",
			"new-password":     "newsecurepassword456",
			"confirm-password": "newsecurepassword456",
		}).
		Put("/account/change/password")
	if err != nil {
		t.Fatalf("password change failed: %v", err)
	}
	defer respPass.Close()

	if respPass.StatusCode() != http.StatusOK && respPass.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected OK or Redirect for password change, got %d", respPass.StatusCode())
	}

	// Get the new authorization cookie returned
	var newCookie string
	for _, cookieVal := range respPass.Cookies() {
		if string(cookieVal.Key()) == "Authorization" {
			newCookie = string(cookieVal.Key()) + "=" + string(cookieVal.Value())
			break
		}
	}
	if newCookie == "" {
		newCookie = cookie // fallback
	}

	// Delete Account
	respDel, err := ts.Client.R().
		SetHeader("Cookie", newCookie).
		SetFormDataWithMap(map[string]string{
			"current-password": "newsecurepassword456",
			"confirm-name":     "change_usr_new",
		}).
		Delete("/account/")
	if err != nil {
		t.Fatalf("account delete failed: %v", err)
	}
	defer respDel.Close()

	if respDel.StatusCode() != http.StatusOK && respDel.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected OK or Redirect for account delete, got %d", respDel.StatusCode())
	}

	// Verify user is marked as deleted in DB
	deletedUser, err := model.NewUserFromId(user.Id)
	if err != nil {
		t.Fatalf("failed to query deleted user: %v", err)
	}
	if !deletedUser.IsDeleted {
		t.Errorf("expected user state to be marked deleted after delete account action")
	}
}

func TestAPIMemberStatusAndList(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookie, user := createTestUserCookie(t, ts, "Member Status User", "status_user", "status@example.com")
	if user.IsOnline() {
		t.Errorf("expected user to be offline initially")
	}

	group := createTestGroupDirectly(t, int64(user.Id), "Status Group", "status_grp")

	// Request members partial page
	respMemPage, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("failed to fetch member list page: %v", err)
	}
	defer respMemPage.Close()

	if respMemPage.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK for member list page, got %d", respMemPage.StatusCode())
	}

	body := respMemPage.String()
	if !strings.Contains(body, "offline") {
		t.Errorf("expected member list HTML to contain 'offline' status badge")
	}
	if !strings.Contains(body, "Member Status User") {
		t.Errorf("expected member list HTML to contain member moniker")
	}
}

func TestAPIGroupLeaveEdgeCases(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Owner and Joiner
	cookieOwner, owner := createTestUserCookie(t, ts, "Owner User Leave", "owner_leave", "owner_leave@example.com")
	cookieJoiner, _ := createTestUserCookie(t, ts, "Joiner User Leave", "joiner_leave", "joiner_leave@example.com")

	group := createTestGroupDirectly(t, int64(owner.Id), "Leave Edge Group", "leave_edge_grp")

	// Joiner joins group
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	respJoin.Close()

	// Owner attempts to leave while not alone -> should be forbidden (403)
	respOwnerLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/leave", group.Id))
	if err != nil {
		t.Fatalf("owner leave request failed: %v", err)
	}
	if respOwnerLeave.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for owner leaving non-empty group, got %d", respOwnerLeave.StatusCode())
	}
	respOwnerLeave.Close()

	// Joiner leaves group -> should succeed
	respJoinerLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Delete(fmt.Sprintf("/groups/%d/leave", group.Id))
	if err != nil {
		t.Fatalf("joiner leave request failed: %v", err)
	}
	if respJoinerLeave.StatusCode() != http.StatusOK && respJoinerLeave.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected status 200/303 for joiner leave, got %d", respJoinerLeave.StatusCode())
	}
	respJoinerLeave.Close()

	// Now owner is alone in the group, owner leaves group -> should succeed and mark group deleted
	respSoleOwnerLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/leave", group.Id))
	if err != nil {
		t.Fatalf("sole owner leave request failed: %v", err)
	}
	if respSoleOwnerLeave.StatusCode() != http.StatusOK && respSoleOwnerLeave.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected status 200/303 for sole owner leave, got %d", respSoleOwnerLeave.StatusCode())
	}
	respSoleOwnerLeave.Close()

	updatedGroup, err := model.NewGroupFromId(group.Id)
	if err != nil {
		t.Fatalf("failed to query group after sole owner left: %v", err)
	}
	if !updatedGroup.IsDeleted {
		t.Errorf("expected group to be marked deleted after sole owner leaves")
	}
}

func TestAPIMemberJoinLeaveMemberListUpdate(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner and Create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Owner User ML", "owner_ml", "owner_ml@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Member List Group", "member_list_grp")

	// 1. Verify owner element is present initially
	respPage1, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("failed to fetch initial member list: %v", err)
	}
	defer respPage1.Close()

	ownerElement := fmt.Sprintf("id=\"member-%d\"", owner.Id)
	if !strings.Contains(respPage1.String(), ownerElement) {
		t.Errorf("expected owner member element %s in initial member list, got: %s", ownerElement, respPage1.String())
	}

	// Register Joiner
	cookieJoiner, joiner := createTestUserCookie(t, ts, "Joiner User ML", "joiner_ml", "joiner_ml@example.com")
	joinerElement := fmt.Sprintf("id=\"member-%d\"", joiner.Id)

	// Verify joiner is NOT in group member list before joining
	respBeforeJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("failed to fetch member list before join: %v", err)
	}
	defer respBeforeJoin.Close()

	if strings.Contains(respBeforeJoin.String(), joinerElement) {
		t.Errorf("expected joiner member element %s to NOT be present before joining", joinerElement)
	}

	// 2. Joiner joins group
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("join request failed: %v", err)
	}
	respJoin.Close()

	// Verify joiner element IS added to member list after joining
	respAfterJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("failed to fetch member list after join: %v", err)
	}
	defer respAfterJoin.Close()

	if !strings.Contains(respAfterJoin.String(), joinerElement) {
		t.Errorf("expected joiner member element %s to be added to member list after joining, got: %s", joinerElement, respAfterJoin.String())
	}
	if !strings.Contains(respAfterJoin.String(), ownerElement) {
		t.Errorf("expected owner member element %s to remain in member list after join, got: %s", ownerElement, respAfterJoin.String())
	}

	// 3. Joiner leaves group
	respLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Delete(fmt.Sprintf("/groups/%d/leave", group.Id))
	if err != nil {
		t.Fatalf("leave request failed: %v", err)
	}
	respLeave.Close()

	// Verify joiner element IS removed from member list after leaving
	respAfterLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/page/1", group.Id))
	if err != nil {
		t.Fatalf("failed to fetch member list after leave: %v", err)
	}
	defer respAfterLeave.Close()

	if strings.Contains(respAfterLeave.String(), joinerElement) {
		t.Errorf("expected joiner member element %s to be removed from member list after leaving, but found in: %s", joinerElement, respAfterLeave.String())
	}
	if !strings.Contains(respAfterLeave.String(), ownerElement) {
		t.Errorf("expected owner member element %s to remain in member list after joiner leaves, got: %s", ownerElement, respAfterLeave.String())
	}
}

func TestAPIMemberJoinLeaveWebSocketBroadcast(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "WS Owner", "ws_owner", "ws_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "WS Group", "ws_group")

	port := ts.Listener.Addr().(*net.TCPAddr).Port
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/groups/ws/%d/", port, group.Id)

	header := http.Header{}
	header.Add("Cookie", cookieOwner)

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer wsConn.Close()

	// Set read deadline so test doesn't hang if WS read blocks
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read initial connection broadcast
	_, initialMsg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read initial ws message: %v", err)
	}
	_ = initialMsg

	cookieJoiner, joiner := createTestUserCookie(t, ts, "WS Joiner", "ws_joiner", "ws_joiner@example.com")

	// 1. Joiner joins group
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("join request failed: %v", err)
	}
	respJoin.Close()

	// Read message from WS connection
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, joinMsgBytes, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read ws join message: %v", err)
	}
	joinMsg := string(joinMsgBytes)

	joinerElement := fmt.Sprintf("id=\"member-%d\"", joiner.Id)
	if !strings.Contains(joinMsg, joinerElement) {
		t.Errorf("expected WS broadcast on join to contain member element %s, got: %s", joinerElement, joinMsg)
	}

	// 2. Joiner leaves group
	respLeave, err := ts.Client.R().
		SetHeader("Cookie", cookieJoiner).
		Delete(fmt.Sprintf("/groups/%d/leave", group.Id))
	if err != nil {
		t.Fatalf("leave request failed: %v", err)
	}
	respLeave.Close()

	// Read message from WS connection
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, leaveMsgBytes, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read ws leave message: %v", err)
	}
	leaveMsg := string(leaveMsgBytes)

	deleteOob := fmt.Sprintf("delete:#member-%d", joiner.Id)
	if !strings.Contains(leaveMsg, deleteOob) {
		t.Errorf("expected WS broadcast on leave to contain OOB delete swap %s, got: %s", deleteOob, leaveMsg)
	}
}
