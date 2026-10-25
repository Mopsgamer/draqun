package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Mopsgamer/draqun/server/model"
	"github.com/golang-jwt/jwt/v5"
)

func TestPermissionsEndpointsAndHTMXFlows(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner User and create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Perm Route Owner", "perm_route_owner", "perm_route_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Perm Endpoints Group", "perm_endpoints_group")

	// Create a text channel
	ch := model.NewChannel("announcements", "Announcements", "Group announcements", model.ChannelTypeText)
	if err := ch.Insert(); err != nil {
		t.Fatalf("failed to insert channel: %v", err)
	}
	if err := model.AddGroupChannel(group.Id, ch.Id, 1); err != nil {
		t.Fatalf("failed to add group channel: %v", err)
	}

	// Create custom role
	role := model.Role{
		GroupId: group.Id,
		Name:    "mod_role",
		Moniker: "Moderator Role",
	}
	if err := role.Insert(); err != nil {
		t.Fatalf("failed to insert role: %v", err)
	}

	// Register Member User
	cookieMember, memberUser := createTestUserCookie(t, ts, "Perm Route Member", "perm_route_member", "perm_route_member@example.com")
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieMember).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("member join failed: %v", err)
	}
	respJoin.Close()

	// 1. GET /groups/:id/settings/permissions -> Permissions Settings Page
	respPermsPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
	if err != nil {
		t.Fatalf("GET permissions page failed: %v", err)
	}
	defer respPermsPage.Close()

	if respPermsPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for permissions page, got %d", respPermsPage.StatusCode())
	}
	if !strings.Contains(respPermsPage.String(), "Group Permissions") {
		t.Errorf("expected permissions page to contain 'Group Permissions', got: %s", respPermsPage.String())
	}

	// 2. POST /groups/:id/permissions -> Create new permission rule
	respCreatePerm, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":     "write",
			"level":    "disable",
			"priority": "10",
		}).
		Post(fmt.Sprintf("/groups/%d/permissions", group.Id))
	if err != nil {
		t.Fatalf("create permission failed: %v", err)
	}
	respCreatePerm.Close()

	if respCreatePerm.StatusCode() != http.StatusOK && respCreatePerm.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for permission creation, got %d", respCreatePerm.StatusCode())
	}

	// Query created permission from DB
	groupPerms, err := model.GetGroupPermissions(group.Id)
	if err != nil || len(groupPerms) != 1 {
		t.Fatalf("expected 1 group permission, got %d (err: %v)", len(groupPerms), err)
	}
	perm1 := groupPerms[0]
	if perm1.Name != model.PermNameWrite || perm1.Level != model.PermLevelDisable {
		t.Errorf("expected write/disable permission, got %s/%s", perm1.Name, perm1.Level)
	}

	// 3. PUT /groups/:id/permissions/:perm_id -> Update permission rule
	respUpdatePerm, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":  "write",
			"level": "all",
		}).
		Put(fmt.Sprintf("/groups/%d/permissions/%d", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("update permission failed: %v", err)
	}
	respUpdatePerm.Close()

	updatedPerm, err := model.NewPermissionFromId(perm1.Id, group.Id)
	if err != nil || updatedPerm.Level != model.PermLevelAll {
		t.Errorf("expected updated permission level 'all', got %s", updatedPerm.Level)
	}

	// 4. Create second rule to test priority reordering
	respCreatePerm2, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":     "read",
			"level":    "own",
			"priority": "20",
		}).
		Post(fmt.Sprintf("/groups/%d/permissions", group.Id))
	if err != nil {
		t.Fatalf("create second permission failed: %v", err)
	}
	respCreatePerm2.Close()

	// Reorder rule
	respMovePerm, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/permissions/%d/move?direction=down", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("move permission failed: %v", err)
	}
	respMovePerm.Close()

	if respMovePerm.StatusCode() != http.StatusOK && respMovePerm.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200/303 for move permission, got %d", respMovePerm.StatusCode())
	}

	// 5. GET /groups/:id/permissions/entities/search -> Search Channels
	respSearchChannels, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/permissions/entities/search?perm_id=%d&entity_type=channel&q=Anno", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("search channels failed: %v", err)
	}
	defer respSearchChannels.Close()

	if respSearchChannels.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for channel search, got %d", respSearchChannels.StatusCode())
	}
	if !strings.Contains(respSearchChannels.String(), "Announcements") {
		t.Errorf("expected channel search to contain 'Announcements', got: %s", respSearchChannels.String())
	}

	// GET /groups/:id/permissions/entities/search -> Search Roles
	respSearchRoles, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/permissions/entities/search?perm_id=%d&entity_type=role&q=Mod", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("search roles failed: %v", err)
	}
	defer respSearchRoles.Close()

	if respSearchRoles.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for role search, got %d", respSearchRoles.StatusCode())
	}
	if !strings.Contains(respSearchRoles.String(), "Moderator Role") {
		t.Errorf("expected role search to contain 'Moderator Role', got: %s", respSearchRoles.String())
	}

	// GET /groups/:id/permissions/entities/search -> Search Users
	respSearchUsers, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/permissions/entities/search?perm_id=%d&entity_type=user&q=Route", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("search users failed: %v", err)
	}
	defer respSearchUsers.Close()

	if respSearchUsers.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for user search, got %d", respSearchUsers.StatusCode())
	}
	if !strings.Contains(respSearchUsers.String(), "Perm Route Member") {
		t.Errorf("expected user search to contain 'Perm Route Member', got: %s", respSearchUsers.String())
	}

	// 6. POST /groups/:id/permissions/:perm_id/attach -> Attach channel, role, user
	respAttachChannel, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"entity_type": "channel",
			"entity_id":   fmt.Sprintf("%d", ch.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/permissions/%d/attach", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("attach channel failed: %v", err)
	}
	respAttachChannel.Close()

	respAttachRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"entity_type": "role",
			"entity_id":   fmt.Sprintf("%d", role.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/permissions/%d/attach", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("attach role failed: %v", err)
	}
	respAttachRole.Close()

	respAttachUser, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"entity_type": "user",
			"entity_id":   fmt.Sprintf("%d", memberUser.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/permissions/%d/attach", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("attach user failed: %v", err)
	}
	respAttachUser.Close()

	// Verify attachments in DB
	permAttached, err := model.NewPermissionFromId(perm1.Id, group.Id)
	if err != nil {
		t.Fatalf("failed to query attached permission: %v", err)
	}
	if len(permAttached.Channels) != 1 || permAttached.Channels[0].Id != ch.Id {
		t.Errorf("expected attached channel %d, got %v", ch.Id, permAttached.Channels)
	}
	if len(permAttached.Roles) != 1 || permAttached.Roles[0].Id != role.Id {
		t.Errorf("expected attached role %d, got %v", role.Id, permAttached.Roles)
	}
	if len(permAttached.Users) != 1 || permAttached.Users[0].Id != memberUser.Id {
		t.Errorf("expected attached user %d, got %v", memberUser.Id, permAttached.Users)
	}

	// 7. DELETE /groups/:id/permissions/:perm_id/detach/:entity_type/:entity_id -> Detach user
	respDetachUser, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/permissions/%d/detach/user/%d", group.Id, perm1.Id, memberUser.Id))
	if err != nil {
		t.Fatalf("detach user failed: %v", err)
	}
	respDetachUser.Close()

	permDetached, err := model.NewPermissionFromId(perm1.Id, group.Id)
	if err != nil {
		t.Fatalf("failed to query detached permission: %v", err)
	}
	if len(permDetached.Users) != 0 {
		t.Errorf("expected 0 attached users after detach, got %d", len(permDetached.Users))
	}

	// 8. DELETE /groups/:id/permissions/:perm_id -> Delete permission rule
	respDeletePerm, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/permissions/%d", group.Id, perm1.Id))
	if err != nil {
		t.Fatalf("delete permission failed: %v", err)
	}
	respDeletePerm.Close()

	permDeleted, err := model.NewPermissionFromId(perm1.Id, group.Id)
	if err == nil && !permDeleted.IsEmpty() {
		t.Errorf("expected permission rule to be deleted, but found in DB")
	}
}

func TestPrivateGroupJoinSecurity(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create owner and private group
	cookieOwner, owner := createTestUserCookie(t, ts, "Private Group Owner", "priv_owner", "priv_owner@example.com")
	_ = cookieOwner
	privGroup := model.NewGroup(owner.Id, "Private Group", "priv_group", model.GroupModePrivate, "", "Private test group", "")
	if err := privGroup.Insert(); err != nil {
		t.Fatalf("failed to create private group: %v", err)
	}

	// Create another user
	cookieAttacker, attacker := createTestUserCookie(t, ts, "Attacker User", "attacker_user", "attacker@example.com")
	_ = attacker

	// Attempt to join private group directly via PUT /groups/:group_id/join
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieAttacker).
		Put(fmt.Sprintf("/groups/%d/join", privGroup.Id))
	if err != nil {
		t.Fatalf("join request failed: %v", err)
	}
	defer respJoin.Close()

	if respJoin.StatusCode() == http.StatusOK {
		t.Errorf("expected error when joining private group, but got 200 OK")
	}

	// Confirm attacker is not a member of private group
	m, errM := model.NewMemberFromId(privGroup.Id, attacker.Id)
	if errM == nil && !m.IsEmpty() && !m.IsDeleted {
		t.Errorf("attacker successfully joined private group without invitation")
	}
}

func TestSearchEndpointsQueryFallbacks(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Search Fallback Owner", "search_owner", "search_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Search Group", "search_group")

	// GET /groups/:id/members/search with query string
	respMemberSearch, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/members/search?q=search_owner", group.Id))
	if err != nil {
		t.Fatalf("member search failed: %v", err)
	}
	defer respMemberSearch.Close()

	if respMemberSearch.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for member search fallback, got %d", respMemberSearch.StatusCode())
	}

	// GET /groups/:id/tickets/search with query string
	respTicketSearch, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/tickets/search?q=test&mode=parent", group.Id))
	if err != nil {
		t.Fatalf("ticket search failed: %v", err)
	}
	defer respTicketSearch.Close()

	if respTicketSearch.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for ticket search fallback, got %d", respTicketSearch.StatusCode())
	}
}

func TestAdminPermissionValidation(t *testing.T) {
	adminPerm := model.Permission{
		GroupId: 1,
		Name:    model.PermNameAdmin,
		Level:   model.PermLevelOwn,
	}
	if alert := adminPerm.Validate(); alert == nil {
		t.Errorf("expected validation failure for Admin perm with 'own' level")
	}

	rolesPerm := model.Permission{
		GroupId: 1,
		Name:    model.PermNameRoles,
		Level:   model.PermLevelOwn,
	}
	if alert := rolesPerm.Validate(); alert == nil {
		t.Errorf("expected validation failure for Roles perm with 'own' level")
	}

	settingsPerm := model.Permission{
		GroupId: 1,
		Name:    model.PermNameSettings,
		Level:   model.PermLevelOwn,
	}
	if alert := settingsPerm.Validate(); alert == nil {
		t.Errorf("expected validation failure for Settings perm with 'own' level")
	}
}

func TestSuspendedUserTokenAuthFailure(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create user and get session cookie
	cookie, user := createTestUserCookie(t, ts, "Suspended User Auth", "susp_user_auth", "susp_auth@example.com")

	// Suspend the user
	user.IsSuspended = true
	if err := user.Update(); err != nil {
		t.Fatalf("failed to update user suspension status: %v", err)
	}

	// Attempt request with existing cookie after suspension
	respList, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		Get("/friends/list")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer respList.Close()

	if respList.StatusCode() == http.StatusOK {
		t.Errorf("expected authentication failure for suspended user, got 200 OK")
	}
}

func TestSlidingTokenRefresh(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create user
	_, user := createTestUserCookie(t, ts, "Sliding Refresh User", "sliding_user", "sliding@example.com")

	// Create a JWT token with expiration time in 1 hour (less than half of lifespan)
	claims := jwt.MapClaims{
		"Id":       float64(user.Id),
		"Password": string(user.Password),
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Add(-23 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("test_jwt_key_that_is_long_enough_12345"))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	cookieVal := fmt.Sprintf("Authorization=%s", tokenStr)

	resp, errReq := ts.Client.R().
		SetHeader("Cookie", cookieVal).
		Get("/friends/list")
	if errReq != nil {
		t.Fatalf("request failed: %v", errReq)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode())
	}

	// Verify that Set-Cookie header was returned with refreshed Authorization token
	cookieHeader := resp.Header("Set-Cookie")
	if !strings.Contains(cookieHeader, "Authorization=") || strings.Contains(cookieHeader, "Authorization=;") {
		t.Errorf("expected Set-Cookie header with refreshed token, got: %s", cookieHeader)
	}
}
