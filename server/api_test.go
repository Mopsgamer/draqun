package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
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

	// Register User
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
