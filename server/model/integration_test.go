package model

import (
	"testing"
    "os"

	"github.com/stretchr/testify/assert"
	"github.com/doug-martin/goqu/v9"
)

func setupDB(t *testing.T) {
    // Force LoadDB to use the absolute path to app_data.db in the root
    // Since LoadDB uses a hardcoded "app_data.db" relative to current working directory,
    // and go test runs in the package directory, we might need to change dir or fix LoadDB.

    // For now, let's try to change the working directory to the root where app_data.db is.
    originalWd, _ := os.Getwd()
    err := os.Chdir("../..")
    if err != nil {
        t.Fatalf("failed to change dir: %v", err)
    }
    t.Cleanup(func() { os.Chdir(originalWd) })

	err = LoadDB()
	if err != nil {
		t.Fatalf("failed to load db: %v", err)
	}
}

func TestMultipleGroupsSameRoleName(t *testing.T) {
	setupDB(t)

	// Create Group 1
	g1 := NewGroup(1, "Group 1", "group1", GroupModePublic, "", "", "")
	err := g1.Insert()
	assert.NoError(t, err)

	// Create Role for Group 1
	r1 := NewRoleEveryone(g1.Id)
	err = r1.Insert()
	assert.NoError(t, err)

	// Create Group 2
	g2 := NewGroup(1, "Group 2", "group2", GroupModePublic, "", "", "")
	err = g2.Insert()
	assert.NoError(t, err)

	// Create Role for Group 2 with same name as Group 1's role
	r2 := NewRoleEveryone(g2.Id)
	err = r2.Insert()
	assert.NoError(t, err, "Should allow same role name in different groups")

    // Cleanup
    _ = r1.Delete()
    _ = r2.Delete()
	_ = Delete(TableGroups, goqu.Ex{"id": g1.Id})
	_ = Delete(TableGroups, goqu.Ex{"id": g2.Id})
}

func TestRoleAssigneeNavigation(t *testing.T) {
	setupDB(t)

	// Setup: Group, User (assumed to exist or we use a dummy ID), Member, Role, RoleAssignee
	g := NewGroup(1, "Nav Group", "navgroup", GroupModePublic, "", "", "")
	_ = g.Insert()

	r := NewRoleEveryone(g.Id)
	r.Name = "Test Role"
	_ = r.Insert()

	ra := RoleAssignee{
		UserId: 1,
		RoleId: r.Id,
	}
	_ = ra.Insert()

	// Test Role()
	fetchedRole := ra.Role()
	assert.Equal(t, r.Id, fetchedRole.Id)
	assert.Equal(t, r.Name, fetchedRole.Name)

	// Test Member()
	// First ensure member exists
	m := NewMember(g.Id, 1, "testmoniker")
	_ = m.Insert()

	fetchedMember := ra.Member()
	assert.Equal(t, g.Id, fetchedMember.GroupId)
	assert.Equal(t, uint64(1), fetchedMember.UserId)

	// Cleanup
	_ = Delete(TableMembers, goqu.Ex{"group_id": g.Id, "user_id": 1})
	_ = ra.Delete()
	_ = r.Delete()
	_ = Delete(TableGroups, goqu.Ex{"id": g.Id})
}
