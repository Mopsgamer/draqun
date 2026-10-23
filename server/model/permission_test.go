package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionCRUDAndEntities(t *testing.T) {
	setupTestDB(t)

	// Create test user and group
	user := NewUser("Perm User", "perm_user", "perm_user@example.com", "1234567890", "hashedpassword123", "")
	err := user.Insert()
	assert.NoError(t, err)

	group := NewGroup(user.Id, "Perm Group", "perm_group", GroupModePublic, "", "Desc", "")
	err = group.Insert()
	assert.NoError(t, err)

	channel := NewChannel("general", "General", "Desc", ChannelTypeText)
	err = channel.Insert()
	assert.NoError(t, err)

	role := Role{
		GroupId: group.Id,
		Name:    "moderator",
		Moniker: "Moderator",
		Color:   0xFF0000,
	}
	err = role.Insert()
	assert.NoError(t, err)

	// 1. Create permission
	perm := Permission{
		GroupId:  group.Id,
		Name:     PermNameWrite,
		Level:    PermLevelAll,
		Priority: 10,
	}
	err = perm.Validate()
	assert.Nil(t, err)

	err = perm.Insert()
	assert.NoError(t, err)
	assert.NotZero(t, perm.Id)
	assert.False(t, perm.IsEmpty())

	// 2. Attach channel, role, user
	err = perm.AddChannel(channel.Id)
	assert.NoError(t, err)

	err = perm.AddRole(role.Id)
	assert.NoError(t, err)

	err = perm.AddUser(user.Id)
	assert.NoError(t, err)

	// Load entities
	fetched, err := NewPermissionFromId(perm.Id, group.Id)
	assert.NoError(t, err)
	assert.False(t, fetched.IsEmpty())
	assert.Len(t, fetched.Channels, 1)
	assert.Equal(t, channel.Id, fetched.Channels[0].Id)
	assert.Len(t, fetched.Roles, 1)
	assert.Equal(t, role.Id, fetched.Roles[0].Id)
	assert.Len(t, fetched.Users, 1)
	assert.Equal(t, user.Id, fetched.Users[0].Id)

	// 3. Update permission
	fetched.Priority = 20
	err = fetched.Update()
	assert.NoError(t, err)

	updated, err := NewPermissionFromId(perm.Id, group.Id)
	assert.NoError(t, err)
	assert.Equal(t, 20, updated.Priority)

	// 4. Detach entity
	err = updated.RemoveChannel(channel.Id)
	assert.NoError(t, err)

	reloaded, err := NewPermissionFromId(perm.Id, group.Id)
	assert.NoError(t, err)
	assert.Len(t, reloaded.Channels, 0)
	assert.Len(t, reloaded.Roles, 1)

	// 5. Query group permissions
	groupPerms, err := GetGroupPermissions(group.Id)
	assert.NoError(t, err)
	assert.Len(t, groupPerms, 1)

	// 6. Delete permission
	err = reloaded.Delete()
	assert.NoError(t, err)

	deleted, err := NewPermissionFromId(perm.Id, group.Id)
	assert.Error(t, err)
	assert.True(t, deleted.IsEmpty())
}

func TestGetUserPermissionLevel(t *testing.T) {
	setupTestDB(t)

	owner := NewUser("Owner User", "owner_user", "owner@example.com", "1234567891", "hashedpassword123", "")
	_ = owner.Insert()

	memberUser := NewUser("Member User", "member_user", "member@example.com", "1234567892", "hashedpassword123", "")
	_ = memberUser.Insert()

	group := NewGroup(owner.Id, "Evaluation Group", "eval_group", GroupModePublic, "", "Desc", "")
	_ = group.Insert()

	member := NewMember(group.Id, memberUser.Id, "Member Moniker")
	_ = member.Insert()

	channel := NewChannel("announcements", "Announcements", "Desc", ChannelTypeText)
	_ = channel.Insert()

	// 1. Owner always gets PermLevelAll
	ownerLevel := GetUserPermissionLevel(group.Id, owner.Id, 0, PermNameWrite)
	assert.Equal(t, PermLevelAll, ownerLevel)

	// 2. Member standard fallback default -> PermLevelOwn
	defaultWriteLevel := GetUserPermissionLevel(group.Id, memberUser.Id, 0, PermNameWrite)
	assert.Equal(t, PermLevelOwn, defaultWriteLevel)

	// 3. Member administrative fallback default -> PermLevelDisable
	defaultAdminLevel := GetUserPermissionLevel(group.Id, memberUser.Id, 0, PermNameAdmin)
	assert.Equal(t, PermLevelDisable, defaultAdminLevel)

	// 4. Grant Admin permission to member -> all permissions become PermLevelAll
	adminPerm := Permission{
		GroupId:  group.Id,
		Name:     PermNameAdmin,
		Level:    PermLevelAll,
		Priority: 100,
	}
	_ = adminPerm.Insert()
	_ = adminPerm.AddUser(memberUser.Id)

	writeWithAdmin := GetUserPermissionLevel(group.Id, memberUser.Id, 0, PermNameWrite)
	assert.Equal(t, PermLevelAll, writeWithAdmin)

	// Remove admin perm attachment
	_ = adminPerm.RemoveUser(memberUser.Id)

	// 5. Test specific channel level rule with higher priority overriding lower priority
	lowPerm := Permission{
		GroupId:  group.Id,
		Name:     PermNameWrite,
		Level:    PermLevelDisable,
		Priority: 5,
	}
	_ = lowPerm.Insert()
	_ = lowPerm.AddUser(memberUser.Id)

	highPerm := Permission{
		GroupId:  group.Id,
		Name:     PermNameWrite,
		Level:    PermLevelAll,
		Priority: 50,
	}
	_ = highPerm.Insert()
	_ = highPerm.AddChannel(channel.Id)

	// Evaluation for channel -> highPerm (Priority 50) wins -> PermLevelAll
	chanLevel := GetUserPermissionLevel(group.Id, memberUser.Id, channel.Id, PermNameWrite)
	assert.Equal(t, PermLevelAll, chanLevel)

	// Evaluation without channel -> lowPerm (Priority 5) wins -> PermLevelDisable
	noChanLevel := GetUserPermissionLevel(group.Id, memberUser.Id, 0, PermNameWrite)
	assert.Equal(t, PermLevelDisable, noChanLevel)
}
