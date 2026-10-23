package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
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
