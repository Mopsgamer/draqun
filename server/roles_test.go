package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestRolesManagementAndModerationPages(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner User and create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Owner User", "role_owner", "role_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Roles Test Group", "roles_test_group")

	// Register Second User (Member/Target)
	cookieMember, memberUser := createTestUserCookie(t, ts, "Member User", "role_member", "role_member@example.com")

	// Member joins group
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieMember).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("member join failed: %v", err)
	}
	respJoin.Close()

	// Post 2 messages as memberUser
	for i := 1; i <= 2; i++ {
		respMsg, err := ts.Client.R().
			SetHeader("Cookie", cookieMember).
			SetFormDataWithMap(map[string]string{
				"content": fmt.Sprintf("Test message %d", i),
			}).
			Post(fmt.Sprintf("/groups/%d/messages/create", group.Id))
		if err != nil {
			t.Fatalf("post message failed: %v", err)
		}
		respMsg.Close()
	}

	// Verify member's MessagesCount()
	mForCount, err := model.NewMemberFromId(group.Id, memberUser.Id)
	if err != nil {
		t.Fatalf("failed to query member: %v", err)
	}
	if count := mForCount.MessagesCount(); count != 2 {
		t.Errorf("expected member message count to be 2, got %d", count)
	}

	// 1. GET /groups/:id/settings/roles
	respRolesPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/roles", group.Id))
	if err != nil {
		t.Fatalf("GET roles page failed: %v", err)
	}
	defer respRolesPage.Close()

	if respRolesPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for roles page, got %d", respRolesPage.StatusCode())
	}
	if !strings.Contains(respRolesPage.String(), "everyone") {
		t.Errorf("expected roles page to contain @everyone role, got: %s", respRolesPage.String())
	}

	// 2. GET /groups/:id/settings/roles/create
	respCreatePage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/roles/create", group.Id))
	if err != nil {
		t.Fatalf("GET role create page failed: %v", err)
	}
	defer respCreatePage.Close()

	if respCreatePage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for role create page, got %d", respCreatePage.StatusCode())
	}

	// 3. POST /groups/:id/roles/create -> Create Moderator role
	respCreateRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":              "moderator",
			"moniker":           "Moderator",
			"color":             "#ff5500",
			"perm_messages":     "delete",
			"perm_roles":        "disallow",
			"perm_members":      "delete",
			"perm_group_change": "disallow",
			"perm_admin":        "disallow",
		}).
		Post(fmt.Sprintf("/groups/%d/roles/create", group.Id))
	if err != nil {
		t.Fatalf("create role request failed: %v", err)
	}
	respCreateRole.Close()

	if respCreateRole.StatusCode() != http.StatusOK && respCreateRole.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for role creation, got %d", respCreateRole.StatusCode())
	}

	// Verify role was created in DB
	modRole, err := model.NewRoleFromName("moderator", group.Id)
	if err != nil || modRole.IsEmpty() {
		t.Fatalf("failed to retrieve created moderator role from DB: %v", err)
	}
	if modRole.Moniker != "Moderator" {
		t.Errorf("expected role moniker Moderator, got %s", modRole.Moniker)
	}

	// 4. GET /groups/:id/settings/roles/:role_id
	respEditRolePage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/roles/%d", group.Id, modRole.Id))
	if err != nil {
		t.Fatalf("GET edit role page failed: %v", err)
	}
	defer respEditRolePage.Close()

	if respEditRolePage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for edit role page, got %d", respEditRolePage.StatusCode())
	}

	// 5. PUT /groups/:id/roles/:role_id -> Update Moderator role
	respUpdateRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":              "moderator",
			"moniker":           "Senior Moderator",
			"color":             "#00ffaa",
			"perm_messages":     "delete",
			"perm_roles":        "allow",
			"perm_members":      "delete",
			"perm_group_change": "disallow",
			"perm_admin":        "disallow",
		}).
		Put(fmt.Sprintf("/groups/%d/roles/%d", group.Id, modRole.Id))
	if err != nil {
		t.Fatalf("update role request failed: %v", err)
	}
	respUpdateRole.Close()

	if respUpdateRole.StatusCode() != http.StatusOK && respUpdateRole.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for role update, got %d", respUpdateRole.StatusCode())
	}

	updatedModRole, err := model.NewRoleFromId(modRole.Id, group.Id)
	if err != nil || updatedModRole.Moniker != "Senior Moderator" {
		t.Errorf("expected updated role moniker Senior Moderator, got %s", updatedModRole.Moniker)
	}

	// Test Role Reordering
	respMoveRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/roles/%d/move?direction=down", group.Id, modRole.Id))
	if err != nil {
		t.Fatalf("move role request failed: %v", err)
	}
	respMoveRole.Close()

	if respMoveRole.StatusCode() != http.StatusOK && respMoveRole.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for role move action, got %d", respMoveRole.StatusCode())
	}

	// 6. POST /groups/:id/members/:user_id/roles -> Assign Senior Moderator role to memberUser
	respAssignRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"role_id": fmt.Sprintf("%d", modRole.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/roles", group.Id, memberUser.Id))
	if err != nil {
		t.Fatalf("assign role request failed: %v", err)
	}
	respAssignRole.Close()

	if respAssignRole.StatusCode() != http.StatusOK && respAssignRole.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for role assignment, got %d", respAssignRole.StatusCode())
	}

	// Verify member has assigned role
	memberObj, err := model.NewMemberFromId(group.Id, memberUser.Id)
	if err != nil {
		t.Fatalf("failed to query member: %v", err)
	}
	roles := memberObj.Roles()
	if len(roles) != 1 || roles[0].Id != modRole.Id {
		t.Errorf("expected member to have role ID %d assigned, got %v", modRole.Id, roles)
	}

	// 7. GET /groups/:id/settings/members -> Verify members page renders
	respMembersPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/members", group.Id))
	if err != nil {
		t.Fatalf("GET members page failed: %v", err)
	}
	defer respMembersPage.Close()

	if respMembersPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for members page, got %d", respMembersPage.StatusCode())
	}
	if !strings.Contains(respMembersPage.String(), "Senior Moderator") {
		t.Errorf("expected members page to contain assigned role Senior Moderator, got: %s", respMembersPage.String())
	}

	// 8. DELETE /groups/:id/members/:user_id/roles/:role_id -> Unassign role
	respUnassignRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/members/%d/roles/%d", group.Id, memberUser.Id, modRole.Id))
	if err != nil {
		t.Fatalf("unassign role request failed: %v", err)
	}
	respUnassignRole.Close()

	if respUnassignRole.StatusCode() != http.StatusOK && respUnassignRole.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for role unassignment, got %d", respUnassignRole.StatusCode())
	}

	rolesAfterUnassign := memberObj.Roles()
	if len(rolesAfterUnassign) != 0 {
		t.Errorf("expected 0 roles assigned after unassignment, got %d", len(rolesAfterUnassign))
	}

	// 9. DELETE /groups/:id/roles/:role_id -> Delete custom role
	respDeleteRole, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/roles/%d", group.Id, modRole.Id))
	if err != nil {
		t.Fatalf("delete role request failed: %v", err)
	}
	respDeleteRole.Close()

	if respDeleteRole.StatusCode() != http.StatusOK && respDeleteRole.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for role deletion, got %d", respDeleteRole.StatusCode())
	}

	// 10. POST /groups/:id/members/:user_id/kick -> Kick member
	respKick, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"description": "Violation of group rules",
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/kick", group.Id, memberUser.Id))
	if err != nil {
		t.Fatalf("kick request failed: %v", err)
	}
	respKick.Close()

	if respKick.StatusCode() != http.StatusOK && respKick.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for kick action, got %d", respKick.StatusCode())
	}

	// Member joins back
	respRejoin, err := ts.Client.R().
		SetHeader("Cookie", cookieMember).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("rejoin failed: %v", err)
	}
	respRejoin.Close()

	// 11. POST /groups/:id/members/:user_id/ban -> Ban member
	respBan, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"description":   "Repeated rule violations",
			"duration_days": "30",
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/ban", group.Id, memberUser.Id))
	if err != nil {
		t.Fatalf("ban request failed: %v", err)
	}
	respBan.Close()

	if respBan.StatusCode() != http.StatusOK && respBan.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for ban action, got %d", respBan.StatusCode())
	}

	// Banned user attempts to join group -> Should fail (403 Forbidden)
	respBannedJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieMember).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("banned join request failed: %v", err)
	}
	if respBannedJoin.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for banned user attempting to join group, got %d", respBannedJoin.StatusCode())
	}
	respBannedJoin.Close()

	// 12. GET /groups/:id/settings/audit -> Verify audit page
	respAuditPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/audit", group.Id))
	if err != nil {
		t.Fatalf("GET audit page failed: %v", err)
	}
	defer respAuditPage.Close()

	if respAuditPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for audit page, got %d", respAuditPage.StatusCode())
	}
	auditBody := respAuditPage.String()
	if !strings.Contains(auditBody, "Ban") || !strings.Contains(auditBody, "Kick") || !strings.Contains(auditBody, "Role Action") {
		t.Errorf("expected audit page to contain Ban, Kick, and Role Action log entries, got: %s", auditBody)
	}

	// 13. POST /groups/:id/members/:user_id/unban -> Unban member
	respUnban, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/members/%d/unban", group.Id, memberUser.Id))
	if err != nil {
		t.Fatalf("unban request failed: %v", err)
	}
	respUnban.Close()

	if respUnban.StatusCode() != http.StatusOK && respUnban.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200 or 303 for unban action, got %d", respUnban.StatusCode())
	}

	// Unbanned member attempts to join group again -> Should succeed!
	respUnbannedJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieMember).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("unbanned join request failed: %v", err)
	}
	respUnbannedJoin.Close()

	if respUnbannedJoin.StatusCode() != http.StatusOK && respUnbannedJoin.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected status 200/303 for unbanned user joining group, got %d", respUnbannedJoin.StatusCode())
	}
}
