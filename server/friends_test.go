package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestFriendRequestsAndFriendList(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// 1. Create test users
	cookie1, user1 := createTestUserCookie(t, ts, "Alice Tester", "alice_test", "alice@example.com")
	cookie2, user2 := createTestUserCookie(t, ts, "Bob Tester", "bob_test", "bob@example.com")
	cookie3, user3 := createTestUserCookie(t, ts, "Charlie Tester", "charlie_test", "charlie@example.com")

	// 2. Test self friend request prevention
	respSelf, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"username": string(user1.Name),
		}).
		Post("/friends/request")
	if err != nil || !strings.Contains(respSelf.String(), "cannot send a friend request to yourself") {
		t.Fatalf("expected error when sending friend request to oneself, got: %v (body: %s)", respSelf.StatusCode(), respSelf.String())
	}

	// 3. Alice sends friend request to Bob by username (@bob_test)
	respReq, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"username": "@bob_test",
		}).
		Post("/friends/request")
	if err != nil || respReq.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from sending friend request, got status: %v", respReq.StatusCode())
	}
	if !strings.Contains(respReq.String(), "Friend request sent to Bob Tester!") {
		t.Fatalf("expected success message in response, got: %s", respReq.String())
	}

	// Verify friendship model state
	friendship, err := model.GetFriendship(user1.Id, user2.Id)
	if err != nil || friendship.IsEmpty() || friendship.Status != model.FriendStatusPending {
		t.Fatalf("expected pending friendship between user1 and user2, got: %v (err: %v)", friendship, err)
	}

	// Verify Bob received a friend_request notification
	notifsBob, err := model.GetUserNotifications(user2.Id, 10)
	if err != nil || len(notifsBob) != 1 || notifsBob[0].Type != "friend_request" {
		t.Fatalf("expected 1 friend_request notification for Bob, got: %v (err: %v)", notifsBob, err)
	}

	// 4. Test duplicate request prevention
	respDup, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"target_user_id": fmt.Sprintf("%d", user2.Id),
		}).
		Post("/friends/request")
	if err != nil || !strings.Contains(respDup.String(), "friend request already sent") {
		t.Fatalf("expected error on duplicate request, got body: %s", respDup.String())
	}

	// 5. Bob accepts Alice's friend request via notification accept endpoint
	notifId := notifsBob[0].Id
	respAccept, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/friends/requests/%d/accept", notifId))
	if err != nil || respAccept.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK when Bob accepts friend request, got status: %v", respAccept.StatusCode())
	}

	// Verify friendship status updated to accepted
	friendshipAfter, err := model.GetFriendship(user1.Id, user2.Id)
	if err != nil || friendshipAfter.Status != model.FriendStatusAccepted {
		t.Fatalf("expected accepted friendship status, got: %v", friendshipAfter)
	}

	// Verify Alice received friend_request_accepted notification
	notifsAlice, err := model.GetUserNotifications(user1.Id, 10)
	if err != nil || len(notifsAlice) != 1 || notifsAlice[0].Type != "friend_request_accepted" {
		t.Fatalf("expected friend_request_accepted notification for Alice, got: %v", notifsAlice)
	}

	// Verify User.Friends() for Alice and Bob
	aliceFriends := user1.Friends()
	if len(aliceFriends) != 1 || aliceFriends[0].Id != user2.Id {
		t.Fatalf("expected Bob in Alice's friends list, got: %v", aliceFriends)
	}

	bobFriends := user2.Friends()
	if len(bobFriends) != 1 || bobFriends[0].Id != user1.Id {
		t.Fatalf("expected Alice in Bob's friends list, got: %v", bobFriends)
	}

	// 6. Test GET /friends/list
	respList, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		Get("/friends/list")
	if err != nil || respList.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from GET /friends/list, got: %v", respList.StatusCode())
	}
	if !strings.Contains(respList.String(), string(user2.Moniker)) {
		t.Fatalf("expected friends list HTML to contain Bob's moniker, got: %s", respList.String())
	}

	// 7. Test declining friend request
	// Charlie sends friend request to Bob
	_, errReq3 := model.SendFriendRequest(user3.Id, user2.Id)
	if errReq3 != nil {
		t.Fatalf("failed sending request from Charlie to Bob: %v", errReq3)
	}
	notifReq3 := model.NewNotification(user2.Id, user3.Id, 0, 0, 0, "friend_request", "Friend Request", "Charlie sent request")
	_ = notifReq3.Insert()

	// Bob declines Charlie's request via POST /friends/requests/:id/decline
	respDecline, err := ts.Client.R().
		SetHeader("Cookie", cookie2).
		SetHeader("HX-Request", "true").
		Post(fmt.Sprintf("/friends/requests/%d/decline", notifReq3.Id))
	if err != nil || respDecline.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from declining request, got: %v", respDecline.StatusCode())
	}

	fDeclined, _ := model.GetFriendship(user3.Id, user2.Id)
	if fDeclined != nil && !fDeclined.IsEmpty() {
		t.Fatalf("expected friendship record to be deleted after decline, got: %v", fDeclined)
	}

	// 8. Test removing a friend via DELETE /friends/:target_id
	respRemove, err := ts.Client.R().
		SetHeader("Cookie", cookie1).
		SetHeader("HX-Request", "true").
		Delete(fmt.Sprintf("/friends/%d", user2.Id))
	if err != nil || respRemove.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 OK from DELETE /friends/:target_id, got: %v", respRemove.StatusCode())
	}

	aliceFriendsAfterRemove := user1.Friends()
	if len(aliceFriendsAfterRemove) != 0 {
		t.Fatalf("expected 0 friends for Alice after removal, got: %d", len(aliceFriendsAfterRemove))
	}

	_ = cookie3
}
