package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/session"
)

func TestCollabVirtualSubchannels(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Collab Owner", "collab_owner", "collab_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Collab Group", "collab_group")

	// 1. Create a collab channel
	respCreateCh, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "collab_ch",
			"moniker":     "Collaborative Lounge",
			"description": "Collaborative Lounge Channel",
			"type":        "collab",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create collab channel failed: %v", err)
	}
	respCreateCh.Close()

	channels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(channels) == 0 {
		t.Fatalf("expected collab channel, got %d (err: %v)", len(channels), err)
	}
	var collabCh *model.Channel
	for _, c := range channels {
		if c.IsCollab() {
			collabCh = &c
			break
		}
	}
	if collabCh == nil {
		t.Fatalf("collab channel was not found")
	}

	// 2. Joining collab channel with subchannel=new should allocate room 1
	respJoinNew, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/collab/channels/%d/join?subchannel=new", group.Id, collabCh.Id))
	if err != nil {
		t.Fatalf("join subchannel new failed: %v", err)
	}
	defer respJoinNew.Close()

	if respJoinNew.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for collab join, got %d", respJoinNew.StatusCode())
	}
	bodyJoin := respJoinNew.String()
	if !strings.Contains(bodyJoin, "Collaboration") {
		t.Errorf("expected rendered Collaboration in join response, got: %s", bodyJoin)
	}

	activeRooms := session.CollabHub.GetActiveSubchannels(collabCh.Id)
	if len(activeRooms) != 1 || activeRooms[0] != "1" {
		t.Fatalf("expected active room '1', got: %v", activeRooms)
	}

	// 3. User 2 joins subchannel=new, should allocate room 2 (since room 1 has active participants)
	cookieUser2, user2 := createTestUserCookie(t, ts, "Collab User 2", "collab_user2", "collab_user2@example.com")
	_, errJoinGrp := ts.Client.R().
		SetHeader("Cookie", cookieUser2).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if errJoinGrp != nil {
		t.Fatalf("user 2 join group failed: %v", errJoinGrp)
	}

	respJoinUser2, err := ts.Client.R().
		SetHeader("Cookie", cookieUser2).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/collab/channels/%d/join?subchannel=new", group.Id, collabCh.Id))
	if err != nil {
		t.Fatalf("user 2 join new failed: %v", err)
	}
	respJoinUser2.Close()

	activeRoomsAfter2 := session.CollabHub.GetActiveSubchannels(collabCh.Id)
	if len(activeRoomsAfter2) != 2 || activeRoomsAfter2[0] != "1" || activeRoomsAfter2[1] != "2" {
		t.Fatalf("expected active rooms ['1', '2'], got: %v", activeRoomsAfter2)
	}

	// 4. Test page chat rendering shows subchannels numerically ("1", "2") without "ROOMS:"
	respPage, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d", group.Id))
	if err != nil {
		t.Fatalf("GET /chat/groups failed: %v", err)
	}
	defer respPage.Close()

	pageHtml := respPage.String()
	if strings.Contains(pageHtml, "ROOMS:") {
		t.Errorf("expected page NOT to contain 'ROOMS:' section, got: %s", pageHtml)
	}
	if !strings.Contains(pageHtml, "/groups/"+fmt.Sprintf("%d", group.Id)+"/collab/channels/"+fmt.Sprintf("%d", collabCh.Id)+"/join?subchannel=1") {
		t.Errorf("expected page to contain subchannel 1 join link")
	}
	if !strings.Contains(pageHtml, "/groups/"+fmt.Sprintf("%d", group.Id)+"/collab/channels/"+fmt.Sprintf("%d", collabCh.Id)+"/join?subchannel=2") {
		t.Errorf("expected page to contain subchannel 2 join link")
	}

	// 5. Test Quick Controls: mic, output, video toggles, bitrate, disconnect
	respToggleInput, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/collab/toggle-input", group.Id))
	if err != nil || respToggleInput.StatusCode() != http.StatusOK {
		t.Fatalf("toggle input failed: %v", err)
	}
	respToggleInput.Close()

	userSess := session.CollabHub.GetUserSession(owner.Id)
	if userSess == nil || !userSess.GlobalInputMuted {
		t.Errorf("expected owner global input to be muted")
	}

	respToggleOutput, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/collab/toggle-output", group.Id))
	if err != nil || respToggleOutput.StatusCode() != http.StatusOK {
		t.Fatalf("toggle output failed: %v", err)
	}
	respToggleOutput.Close()

	if userSess == nil || !userSess.GlobalOutputMuted {
		t.Errorf("expected owner global output to be muted")
	}

	respToggleVideo, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/collab/toggle-video", group.Id))
	if err != nil || respToggleVideo.StatusCode() != http.StatusOK {
		t.Fatalf("toggle video failed: %v", err)
	}
	respToggleVideo.Close()

	if userSess == nil || !userSess.IsCameraEnabled {
		t.Errorf("expected owner camera to be enabled")
	}

	respDisconnect, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/groups/%d/collab/disconnect", group.Id))
	if err != nil || respDisconnect.StatusCode() != http.StatusOK {
		t.Fatalf("disconnect failed: %v", err)
	}
	respDisconnect.Close()

	userSessAfterDisc := session.CollabHub.GetUserSession(owner.Id)
	if userSessAfterDisc != nil {
		t.Errorf("expected user session to be nil after disconnect")
	}

	_ = user2
}
