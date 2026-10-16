package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestChannelsManagementAndMessaging(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner User and create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Channel Owner", "ch_owner", "ch_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Channel Test Group", "channel_test_group")

	// Verify group has no channels by default
	channels, err := model.GetGroupChannels(group.Id)
	if err != nil {
		t.Fatalf("failed to query group channels: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("expected group to have 0 channels on creation by default, got %d", len(channels))
	}

	// 1. GET /groups/:id/settings/channels
	respChannelsPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/channels", group.Id))
	if err != nil {
		t.Fatalf("GET channels settings page failed: %v", err)
	}
	defer respChannelsPage.Close()

	if respChannelsPage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for channels settings page, got %d", respChannelsPage.StatusCode())
	}
	if !strings.Contains(respChannelsPage.String(), "Group Channels") {
		t.Errorf("expected channels page content, got: %s", respChannelsPage.String())
	}

	// 1b. GET /groups/:id/settings/channels/create
	respCreatePage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/settings/channels/create", group.Id))
	if err != nil {
		t.Fatalf("GET channel create page failed: %v", err)
	}
	defer respCreatePage.Close()

	if respCreatePage.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for channel create page, got %d", respCreatePage.StatusCode())
	}
	if !strings.Contains(respCreatePage.String(), "Create Channel") {
		t.Errorf("expected channel create page content, got: %s", respCreatePage.String())
	}

	// 2. POST /groups/:id/channels/create -> Create main channel
	respCreateMain, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "general",
			"moniker":     "General Chat",
			"description": "General text chat",
			"type":        "text",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create channel failed: %v", err)
	}
	respCreateMain.Close()

	if respCreateMain.StatusCode() != http.StatusOK && respCreateMain.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200/303 for channel creation, got %d", respCreateMain.StatusCode())
	}

	// Verify channel was created
	channels, err = model.GetGroupChannels(group.Id)
	if err != nil || len(channels) != 1 {
		t.Fatalf("expected 1 group channel, got %d (err: %v)", len(channels), err)
	}
	mainChannel := channels[0]
	if mainChannel.Moniker != "General Chat" {
		t.Errorf("expected moniker 'General Chat', got '%s'", mainChannel.Moniker)
	}

	// 3. POST /groups/:id/channels/create -> Create subchannel
	respCreateSub, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "announcements",
			"moniker":     "Announcements Subchannel",
			"description": "Important announcements",
			"type":        "text",
			"parent_id":   fmt.Sprintf("%d", mainChannel.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create subchannel failed: %v", err)
	}
	respCreateSub.Close()

	subchannels, err := model.GetSubchannels(mainChannel.Id)
	if err != nil || len(subchannels) != 1 {
		t.Fatalf("expected 1 subchannel under mainChannel, got %d (err: %v)", len(subchannels), err)
	}
	subChannel := subchannels[0]
	if subChannel.Moniker != "Announcements Subchannel" {
		t.Errorf("expected subchannel moniker 'Announcements Subchannel', got '%s'", subChannel.Moniker)
	}

	// 3b. Create another top-level channel to test moving/reordering
	respCreateSecond, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "random",
			"moniker":     "Random Chat",
			"description": "Random discussion",
			"type":        "text",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create second channel failed: %v", err)
	}
	respCreateSecond.Close()

	topChannels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(topChannels) != 2 {
		t.Fatalf("expected 2 top-level channels, got %d", len(topChannels))
	}
	if topChannels[0].Id != mainChannel.Id {
		t.Errorf("expected first channel to be mainChannel (%d), got %d", mainChannel.Id, topChannels[0].Id)
	}

	// Move mainChannel down
	respMoveDown, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/channels/%d/move?direction=down", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("move channel down failed: %v", err)
	}
	respMoveDown.Close()

	topChannelsAfterMove, _ := model.GetGroupChannels(group.Id)
	if len(topChannelsAfterMove) == 2 && topChannelsAfterMove[0].Id == mainChannel.Id {
		t.Errorf("expected mainChannel to move down to index 1, but it's still at index 0")
	}

	// Move mainChannel back up
	respMoveUp, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Post(fmt.Sprintf("/groups/%d/channels/%d/move?direction=up", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("move channel up failed: %v", err)
	}
	respMoveUp.Close()

	topChannelsBackUp, _ := model.GetGroupChannels(group.Id)
	if len(topChannelsBackUp) == 2 && topChannelsBackUp[0].Id != mainChannel.Id {
		t.Errorf("expected mainChannel to move back up to index 0")
	}

	// 4. POST /groups/:id/channels/:channel_id/messages/create -> Post message to main channel
	respMsg, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"content": "Hello from General Chat!",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("post message to channel failed: %v", err)
	}
	respMsg.Close()

	if respMsg.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for message creation in channel, got %d", respMsg.StatusCode())
	}

	// Verify message in channel
	msgs := mainChannel.MessagesPage(1, 10)
	if len(msgs) != 1 || msgs[0].Content != "Hello from General Chat!" {
		t.Errorf("expected 1 message in channel with content 'Hello from General Chat!', got %v", msgs)
	}

	// 5. GET /groups/:id/channels/:channel_id/messages/page/1 -> Fetch paginated messages
	respGetMsgs, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("GET channel messages page failed: %v", err)
	}
	defer respGetMsgs.Close()

	if respGetMsgs.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for channel messages page, got %d", respGetMsgs.StatusCode())
	}

	// 6. Test Channel Permission Override
	everyoneRole := group.Everyone()
	respSetPerm, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"role_id":       fmt.Sprintf("%d", everyoneRole.Id),
			"perm_messages": "read",
			"inherit":       "false",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/permissions", group.Id, subChannel.Id))
	if err != nil {
		t.Fatalf("set channel permission override failed: %v", err)
	}
	respSetPerm.Close()

	effectiveRole := model.EffectiveRolePermission(subChannel.Id, everyoneRole)
	if effectiveRole.PermMessages != model.PermMessagesRead {
		t.Errorf("expected effective PermMessages to be 'read', got '%s'", effectiveRole.PermMessages)
	}

	// 7. GET /chat/groups/:id/channels/:channel_id -> View channel chat page
	respViewChannel, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, mainChannel.Id))
	if err != nil {
		t.Fatalf("GET chat channel page failed: %v", err)
	}
	defer respViewChannel.Close()

	if respViewChannel.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for chat channel page, got %d", respViewChannel.StatusCode())
	}
	if !strings.Contains(respViewChannel.String(), "General Chat") {
		t.Errorf("expected chat channel page to contain channel moniker 'General Chat', got: %s", respViewChannel.String())
	}

	// 8. DELETE /groups/:id/channels/:channel_id -> Delete subchannel
	respDeleteSub, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/channels/%d", group.Id, subChannel.Id))
	if err != nil {
		t.Fatalf("delete subchannel failed: %v", err)
	}
	respDeleteSub.Close()

	subchannelsAfter, _ := model.GetSubchannels(mainChannel.Id)
	if len(subchannelsAfter) != 0 {
		t.Errorf("expected 0 subchannels after deletion, got %d", len(subchannelsAfter))
	}
}
