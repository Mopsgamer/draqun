package model

import (
	"testing"
)

func TestFirstUserAutomaticDraqunGroupCreation(t *testing.T) {
	setupTestDB(t)

	// Verify no draqun group exists initially
	existingGroup, _ := NewGroupFromName("draqun")
	if !existingGroup.IsEmpty() {
		t.Fatalf("expected no draqun group initially, but found one")
	}

	// Create first user
	user := NewUser("First User", "firstuser", "first@example.com", "1234567890", "password_hash", "")
	if err := user.Insert(); err != nil {
		t.Fatalf("failed to insert first user: %v", err)
	}

	// Fetch draqun group
	draqunGroup, err := NewGroupFromName("draqun")
	if err != nil || draqunGroup.IsEmpty() {
		t.Fatalf("expected draqun group to be automatically created on first user registration, but failed: %v", err)
	}

	if draqunGroup.CreatorId != user.Id {
		t.Errorf("expected draqun group CreatorId to be %d, got %d", user.Id, draqunGroup.CreatorId)
	}

	if draqunGroup.OwnerId != user.Id {
		t.Errorf("expected draqun group OwnerId to be %d, got %d", user.Id, draqunGroup.OwnerId)
	}

	// Verify channels
	channels, err := GetGroupChannels(draqunGroup.Id)
	if err != nil {
		t.Fatalf("failed to get group channels for draqun group: %v", err)
	}

	expectedChannels := map[string]ChannelType{
		"general":         ChannelTypeText,
		"internal":        ChannelTypeText,
		"bugs":            ChannelTypeTicket,
		"feats":           ChannelTypeTicket,
		"admin-requests":  ChannelTypeTicket,
		"private-support": ChannelTypeTicket,
		"public-support":  ChannelTypeTicket,
	}

	if len(channels) != len(expectedChannels) {
		t.Fatalf("expected %d channels, got %d", len(expectedChannels), len(channels))
	}

	channelMap := make(map[string]Channel)
	for _, ch := range channels {
		expectedType, found := expectedChannels[string(ch.Name)]
		if !found {
			t.Errorf("unexpected channel name found: %s", ch.Name)
			continue
		}
		if ch.Type != expectedType {
			t.Errorf("expected channel %s to have type %s, got %s", ch.Name, expectedType, ch.Type)
		}
		channelMap[string(ch.Name)] = ch
	}

	// Verify Roles
	adminRole, err := NewRoleFromName("admin", draqunGroup.Id)
	if err != nil || adminRole.IsEmpty() {
		t.Errorf("expected admin role to exist")
	}

	supportRole, err := NewRoleFromName("support", draqunGroup.Id)
	if err != nil || supportRole.IsEmpty() {
		t.Errorf("expected support role to exist")
	}

	punisherRole, err := NewRoleFromName("punisher", draqunGroup.Id)
	if err != nil || punisherRole.IsEmpty() {
		t.Errorf("expected punisher role to exist")
	}

	// Create test users and assign roles
	regularUser := NewUser("Regular User", "regular", "reg@example.com", "1234567891", "password_hash", "")
	_ = regularUser.Insert()
	regMember := NewMember(draqunGroup.Id, regularUser.Id, "")
	_ = regMember.Insert()

	supportUser := NewUser("Support User", "suppuser", "supp@example.com", "1234567892", "password_hash", "")
	_ = supportUser.Insert()
	suppMember := NewMember(draqunGroup.Id, supportUser.Id, "")
	_ = suppMember.Insert()
	raSupp := RoleAssignee{RoleId: supportRole.Id, UserId: supportUser.Id}
	_ = raSupp.Insert()

	punisherUser := NewUser("Punisher User", "punuser", "pun@example.com", "1234567893", "password_hash", "")
	_ = punisherUser.Insert()
	punMember := NewMember(draqunGroup.Id, punisherUser.Id, "")
	_ = punMember.Insert()
	raPun := RoleAssignee{RoleId: punisherRole.Id, UserId: punisherUser.Id}
	_ = raPun.Insert()

	// Verify channel permission configurations
	internalCh := channelMap["internal"]
	adminReqCh := channelMap["admin-requests"]

	// Internal channel permissions:
	if level := GetUserPermissionLevel(draqunGroup.Id, user.Id, internalCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("owner expected PermLevelAll on internal read, got %s", level)
	}
	if level := GetUserPermissionLevel(draqunGroup.Id, supportUser.Id, internalCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("support expected PermLevelAll on internal read, got %s", level)
	}
	if level := GetUserPermissionLevel(draqunGroup.Id, punisherUser.Id, internalCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("punisher expected PermLevelAll on internal read, got %s", level)
	}
	if level := GetUserPermissionLevel(draqunGroup.Id, regularUser.Id, internalCh.Id, PermNameRead); level != PermLevelDisable {
		t.Errorf("regular user expected PermLevelDisable on internal read, got %s", level)
	}
	privSuppCh := channelMap["private-support"]
	pubSuppCh := channelMap["public-support"]

	// Admin requests permissions:
	// Owner (Admin): Has all access
	if level := GetUserPermissionLevel(draqunGroup.Id, user.Id, adminReqCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("owner expected PermLevelAll on admin-requests read, got %s", level)
	}
	// Support: Read All
	if level := GetUserPermissionLevel(draqunGroup.Id, supportUser.Id, adminReqCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("support expected PermLevelAll on admin-requests read, got %s", level)
	}
	// Punisher: Read All
	if level := GetUserPermissionLevel(draqunGroup.Id, punisherUser.Id, adminReqCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("punisher expected PermLevelAll on admin-requests read, got %s", level)
	}
	// Regular member: Read Disabled
	if level := GetUserPermissionLevel(draqunGroup.Id, regularUser.Id, adminReqCh.Id, PermNameRead); level != PermLevelDisable {
		t.Errorf("regular user expected PermLevelDisable on admin-requests read, got %s", level)
	}

	// Private support permissions:
	// Regular member: Read Own
	if level := GetUserPermissionLevel(draqunGroup.Id, regularUser.Id, privSuppCh.Id, PermNameRead); level != PermLevelOwn {
		t.Errorf("regular user expected PermLevelOwn on private-support read, got %s", level)
	}

	// Public support permissions:
	// Everyone: Read All, Write Disabled, Delete Disabled
	if level := GetUserPermissionLevel(draqunGroup.Id, regularUser.Id, pubSuppCh.Id, PermNameRead); level != PermLevelAll {
		t.Errorf("regular user expected PermLevelAll on public-support read, got %s", level)
	}
	if level := GetUserPermissionLevel(draqunGroup.Id, regularUser.Id, pubSuppCh.Id, PermNameWrite); level != PermLevelDisable {
		t.Errorf("regular user expected PermLevelDisable on public-support write, got %s", level)
	}
	if level := GetUserPermissionLevel(draqunGroup.Id, regularUser.Id, pubSuppCh.Id, PermNameDelete); level != PermLevelDisable {
		t.Errorf("regular user expected PermLevelDisable on public-support delete, got %s", level)
	}
	// Support role: Write All
	if level := GetUserPermissionLevel(draqunGroup.Id, supportUser.Id, pubSuppCh.Id, PermNameWrite); level != PermLevelAll {
		t.Errorf("support user expected PermLevelAll on public-support write, got %s", level)
	}
}
