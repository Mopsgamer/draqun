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
		"general": ChannelTypeText,
		"bugs":    ChannelTypeTicket,
		"feats":   ChannelTypeTicket,
		"reports": ChannelTypeTicket,
	}

	if len(channels) != len(expectedChannels) {
		t.Fatalf("expected %d channels, got %d", len(expectedChannels), len(channels))
	}

	for _, ch := range channels {
		expectedType, found := expectedChannels[string(ch.Name)]
		if !found {
			t.Errorf("unexpected channel name found: %s", ch.Name)
			continue
		}
		if ch.Type != expectedType {
			t.Errorf("expected channel %s to have type %s, got %s", ch.Name, expectedType, ch.Type)
		}
	}
}
