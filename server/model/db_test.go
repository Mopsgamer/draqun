package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Mopsgamer/draqun/server/environment"
)

func chdirToRoot() {
	for i := 0; i < 5; i++ {
		if _, err := os.Stat("go.mod"); err == nil {
			return
		}
		_ = os.Chdir("..")
	}
}

func setupTestDB(t *testing.T) {
	chdirToRoot()

	// Use a unique temporary SQLite file for this test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_model_data.db")

	os.Setenv("JWT_KEY", "test_jwt_key_that_is_long_enough_12345")
	os.Setenv("DB_PATH", dbPath)

	if Sqlx != nil {
		_ = Sqlx.Close()
		Sqlx = nil
	}

	err := environment.LoadEnv(nil)
	if err != nil {
		t.Fatalf("failed to load env: %v", err)
	}

	err = LoadDB()
	if err != nil {
		t.Fatalf("failed to load db: %v", err)
	}

	sqlFileList := []string{
		"./scripts/queries/create_users.sql",
		"./scripts/queries/create_groups.sql",
		"./scripts/queries/create_group_members.sql",
		"./scripts/queries/create_group_roles.sql",
		"./scripts/queries/create_group_role_assignees.sql",
		"./scripts/queries/create_group_messages.sql",
		"./scripts/queries/create_group_action_memberships.sql",
		"./scripts/queries/create_group_action_kicks.sql",
		"./scripts/queries/create_group_action_bans.sql",
	}

	for _, file := range sqlFileList {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read sql file %s: %v", file, err)
		}
		_, err = Sqlx.Exec(string(content))
		if err != nil {
			t.Fatalf("failed to execute sql script %s: %v", file, err)
		}
	}
}

func TestModelUserOperations(t *testing.T) {
	setupTestDB(t)

	// Create user
	user := NewUser("John Doe", "john", "john@example.com", "1234567890", "hashed_password", "")

	// Validation
	if err := user.Validate(); err != nil {
		t.Fatalf("expected user to be valid, got validation error: %v", err)
	}

	// Insert
	if err := user.Insert(); err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	if user.Id == 0 {
		t.Errorf("expected inserted user to have a non-zero ID")
	}

	// Fetch operations
	byEmail, err := NewUserFromEmail("john@example.com")
	if err != nil {
		t.Fatalf("failed to fetch user by email: %v", err)
	}
	if byEmail.Name != "john" {
		t.Errorf("expected name 'john', got '%s'", byEmail.Name)
	}

	byName, err := NewUserFromName("john")
	if err != nil {
		t.Fatalf("failed to fetch user by name: %v", err)
	}
	if byName.Id != user.Id {
		t.Errorf("ID mismatch: %d vs %d", byName.Id, user.Id)
	}

	byId, err := NewUserFromId(user.Id)
	if err != nil {
		t.Fatalf("failed to fetch user by ID: %v", err)
	}
	if byId.Moniker != "John Doe" {
		t.Errorf("moniker mismatch: expected John Doe, got %s", byId.Moniker)
	}

	// Update
	user.Moniker = "John Updated"
	if err := user.Update(); err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	updatedUser, err := NewUserFromId(user.Id)
	if err != nil {
		t.Fatalf("failed to fetch user after update: %v", err)
	}
	if updatedUser.Moniker != "John Updated" {
		t.Errorf("expected moniker to be updated, got '%s'", updatedUser.Moniker)
	}
}

func TestModelGroupOperations(t *testing.T) {
	setupTestDB(t)

	// Create a creator user first
	creator := NewUser("Creator", "creator", "creator@example.com", "1234567891", "password_hash", "")
	_ = creator.Insert()

	// Create Group
	group := NewGroup(creator.Id, "My Test Group", "my_test_group", GroupModePublic, "", "A nice description", "")
	if err := group.Validate(); err != nil {
		t.Fatalf("group validation failed: %v", err)
	}

	if err := group.Insert(); err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}
	if group.Id == 0 {
		t.Errorf("expected group ID to be non-zero after insert")
	}

	// Fetch Group
	byId, err := NewGroupFromId(group.Id)
	if err != nil {
		t.Fatalf("failed to fetch group by ID: %v", err)
	}
	if byId.Name != "my_test_group" {
		t.Errorf("expected group name to be 'my_test_group', got '%s'", byId.Name)
	}

	byName, err := NewGroupFromName("my_test_group")
	if err != nil {
		t.Fatalf("failed to fetch group by name: %v", err)
	}
	if byName.Id != group.Id {
		t.Errorf("expected group ID mismatch, expected %d, got %d", group.Id, byName.Id)
	}

	// Owner/Creator relationships
	ownerUser := group.Owner()
	if ownerUser.Id != creator.Id {
		t.Errorf("expected owner to be creator ID %d, got %d", creator.Id, ownerUser.Id)
	}

	creatorUser := group.Creator()
	if creatorUser.Id != creator.Id {
		t.Errorf("expected creator to be ID %d, got %d", creator.Id, creatorUser.Id)
	}

	// Update Group
	group.Description = "Updated Description"
	if err := group.Update(); err != nil {
		t.Fatalf("failed to update group: %v", err)
	}

	updated, _ := NewGroupFromId(group.Id)
	if updated.Description != "Updated Description" {
		t.Errorf("expected description to be updated, got '%s'", updated.Description)
	}
}

func TestModelMemberOperations(t *testing.T) {
	setupTestDB(t)

	// Setup user and group
	creator := NewUser("Creator", "creator", "creator@example.com", "1234567891", "hash", "")
	_ = creator.Insert()

	group := NewGroup(creator.Id, "Group", "group", GroupModePublic, "", "Description", "")
	_ = group.Insert()

	// Create Member
	member := NewMember(group.Id, creator.Id, "Creator Moniker")
	if err := member.Validate(); err != nil {
		t.Fatalf("member validation failed: %v", err)
	}

	if err := member.Insert(); err != nil {
		t.Fatalf("failed to insert member: %v", err)
	}

	// Fetch Member
	fetched, err := NewMemberFromId(group.Id, creator.Id)
	if err != nil {
		t.Fatalf("failed to fetch member: %v", err)
	}
	if fetched.Moniker != "Creator Moniker" {
		t.Errorf("expected moniker 'Creator Moniker', got '%s'", fetched.Moniker)
	}

	// Check member relationships
	memUser := fetched.User()
	if memUser.Id != creator.Id {
		t.Errorf("member User relationship failed: expected %d, got %d", creator.Id, memUser.Id)
	}

	memGroup := fetched.Group()
	if memGroup.Id != group.Id {
		t.Errorf("member Group relationship failed: expected %d, got %d", group.Id, memGroup.Id)
	}

	// Update Member
	member.Moniker = "New Moniker"
	if err := member.Update(); err != nil {
		t.Fatalf("failed to update member: %v", err)
	}

	updated, _ := NewMemberFromId(group.Id, creator.Id)
	if updated.Moniker != "New Moniker" {
		t.Errorf("expected moniker to be updated, got '%s'", updated.Moniker)
	}
}

func TestModelMessageAndPagination(t *testing.T) {
	setupTestDB(t)

	// Setup user, group, and member
	creator := NewUser("Creator", "creator", "creator@example.com", "1234567891", "hash", "")
	_ = creator.Insert()

	group := NewGroup(creator.Id, "Group", "group", GroupModePublic, "", "Description", "")
	_ = group.Insert()

	member := NewMember(group.Id, creator.Id, "Moniker")
	_ = member.Insert()

	// Insert messages
	for i := 1; i <= 5; i++ {
		msg := NewMessageFilled(group.Id, creator.Id, fmt.Sprintf("Message #%d", i))
		if err := msg.Validate(); err != nil {
			t.Fatalf("message #%d validation failed: %v", i, err)
		}
		if err := msg.Insert(); err != nil {
			t.Fatalf("failed to insert message #%d: %v", i, err)
		}
	}

	// Test Group.MessageFirst and MessageLast
	first := group.MessageFirst()
	if first.Content != "Message #1" {
		t.Errorf("expected first message to be 'Message #1', got '%s'", first.Content)
	}

	last := group.MessageLast()
	if last.Content != "Message #5" {
		t.Errorf("expected last message to be 'Message #5', got '%s'", last.Content)
	}

	// Test Messages Pagination
	page1 := group.MessagesPage(1, 3)
	if len(page1) != 3 {
		t.Errorf("expected 3 messages on page 1, got %d", len(page1))
	}
	// order on pages should be newest first (MessageLast is first in pagination usually), but ordered ascending internally.
	// In group.go: MessagesPage does Order(I("id").Desc()).Limit(limit).Offset(from) in a subquery,
	// and then Order(I("id").Asc()) on top. So it returns the 3 most recent messages ordered ascendingly.
	// Most recent 3 messages: Message #3, Message #4, Message #5.
	if page1[0].Content != "Message #3" || page1[2].Content != "Message #5" {
		t.Errorf("expected message order to be #3 to #5, got %s and %s", page1[0].Content, page1[2].Content)
	}
}

func TestModelRoles(t *testing.T) {
	setupTestDB(t)

	// Setup user and group
	creator := NewUser("Creator", "creator", "creator@example.com", "1234567891", "hash", "")
	_ = creator.Insert()

	group := NewGroup(creator.Id, "Group", "group", GroupModePublic, "", "Description", "")
	_ = group.Insert()

	// Role Everyone
	everyone := NewRoleEveryone(group.Id)
	if err := everyone.Insert(); err != nil {
		t.Fatalf("failed to insert role everyone: %v", err)
	}

	// Role Custom
	custom := Role{
		GroupId: group.Id,
		Name:    "admin",
		Moniker: "Admin",
		PermMessages: PermMessagesDelete,
	}
	if err := custom.Insert(); err != nil {
		t.Fatalf("failed to insert custom role: %v", err)
	}

	// Role Assignee
	assignee := RoleAssignee{
		RoleId: custom.Id,
		UserId: creator.Id,
	}
	if err := assignee.Insert(); err != nil {
		t.Fatalf("failed to insert assignee: %v", err)
	}

	// Fetch Roles
	fetched, err := NewRoleFromName("admin", group.Id)
	if err != nil {
		t.Fatalf("failed to fetch role: %v", err)
	}
	if fetched.PermMessages != PermMessagesDelete {
		t.Errorf("expected PermMessagesDelete, got %v", fetched.PermMessages)
	}
}
