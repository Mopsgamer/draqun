package internal

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
)

func TestTicketChannelAndIssues(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner User and create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Ticket Owner", "ticket_owner", "ticket_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Ticket Test Group", "ticket_test_group")

	// 1. Create top-level Ticket channel
	respCreateTicketCh, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "issues",
			"moniker":     "Bug Tracker",
			"description": "Group issue tracker",
			"type":        "ticket",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create ticket channel failed: %v", err)
	}
	respCreateTicketCh.Close()

	channels, err := model.GetGroupChannels(group.Id)
	if err != nil || len(channels) != 1 {
		t.Fatalf("expected 1 ticket channel, got %d (err: %v)", len(channels), err)
	}
	ticketChannel := channels[0]
	if ticketChannel.Type != model.ChannelTypeTicket {
		t.Errorf("expected channel type 'ticket', got '%s'", ticketChannel.Type)
	}

	// 2. Create Ticket 1 (Subchannel under ticketChannel)
	respCreateIssue1, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":   "Fix login button glitch",
			"content": "The login button overflows on small screens.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("create ticket 1 failed: %v", err)
	}
	respCreateIssue1.Close()

	// Verify Ticket 1 was created as a subchannel
	issues, err := model.GetSubchannels(ticketChannel.Id)
	if err != nil || len(issues) != 1 {
		t.Fatalf("expected 1 issue subchannel, got %d (err: %v)", len(issues), err)
	}
	issue1 := issues[0]
	if issue1.Moniker != "Fix login button glitch" {
		t.Errorf("expected title 'Fix login button glitch', got '%s'", issue1.Moniker)
	}
	if issue1.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", issue1.Status)
	}
	if issue1.AuthorId != owner.Id {
		t.Errorf("expected author ID %d, got %d", owner.Id, issue1.AuthorId)
	}

	// 3. Create Ticket 2 (Subchannel under ticketChannel)
	respCreateIssue2, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":   "Add dark mode support",
			"content": "Users want a dark mode option.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("create ticket 2 failed: %v", err)
	}
	respCreateIssue2.Close()

	issues, _ = model.GetSubchannels(ticketChannel.Id)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issue subchannels, got %d", len(issues))
	}
	issue2 := issues[1]
	if issue2.Id == issue1.Id {
		issue2 = issues[0]
	}

	// 4. Post a comment on Ticket 1
	respComment, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"content": "I am working on this fix right now.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("post comment to issue 1 failed: %v", err)
	}
	respComment.Close()

	if issue1.CommentCount() != 1 {
		t.Errorf("expected comment count 1 for issue 1, got %d", issue1.CommentCount())
	}

	// 5. Close Ticket 1 as Completed
	respClose1, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"status":        "closed",
			"closed_reason": "completed",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/status", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("close issue 1 failed: %v", err)
	}
	respClose1.Close()

	issue1Updated, _ := model.NewChannelFromId(issue1.Id)
	if issue1Updated.Status != "closed" || issue1Updated.ClosedReason != "completed" {
		t.Errorf("expected issue 1 to be closed as completed, got status='%s', reason='%s'", issue1Updated.Status, issue1Updated.ClosedReason)
	}

	// 6. Close Ticket 2 as Duplicate of Ticket 1
	respClose2, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"status":          "closed",
			"closed_reason":   "duplicate",
			"duplicate_of_id": fmt.Sprintf("%d", issue1.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/status", group.Id, issue2.Id))
	if err != nil {
		t.Fatalf("close issue 2 failed: %v", err)
	}
	respClose2.Close()

	issue2Updated, _ := model.NewChannelFromId(issue2.Id)
	if issue2Updated.Status != "closed" || issue2Updated.ClosedReason != "duplicate" || issue2Updated.DuplicateOfId != issue1.Id {
		t.Errorf("expected issue 2 closed as duplicate of %d, got status='%s', reason='%s', dupId=%d", issue1.Id, issue2Updated.Status, issue2Updated.ClosedReason, issue2Updated.DuplicateOfId)
	}

	// 7. Reopen Ticket 1
	respReopen1, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"status": "open",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/status", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("reopen issue 1 failed: %v", err)
	}
	respReopen1.Close()

	issue1Reopened, _ := model.NewChannelFromId(issue1.Id)
	if issue1Reopened.Status != "open" {
		t.Errorf("expected issue 1 to be open, got '%s'", issue1Reopened.Status)
	}

	// 8. Create a sub-ticket under Ticket 1 (sub-task)
	respSubTask, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":   "Update CSS media queries",
			"content": "Adjust breakpoint styles.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("create sub-ticket under issue 1 failed: %v", err)
	}
	respSubTask.Close()

	subTasks, err := model.GetSubchannels(issue1.Id)
	if err != nil || len(subTasks) != 1 {
		t.Fatalf("expected 1 sub-task under issue 1, got %d", len(subTasks))
	}
	subTask := subTasks[0]

	// Verify progress before closing sub-task (0 of 1)
	progressInitial := issue1Reopened.GetTicketProgress()
	if progressInitial.TotalCount != 1 || progressInitial.ClosedCount != 0 || progressInitial.Percentage != 0 {
		t.Errorf("expected initial progress 0/1 (0%%), got %v", progressInitial)
	}

	// Close sub-task
	respCloseSub, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"status":        "closed",
			"closed_reason": "completed",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/status", group.Id, subTask.Id))
	if err != nil {
		t.Fatalf("close sub-task failed: %v", err)
	}
	respCloseSub.Close()

	// Verify progress after closing sub-task (1 of 1, 100%)
	progressClosed := issue1Reopened.GetTicketProgress()
	if progressClosed.TotalCount != 1 || progressClosed.ClosedCount != 1 || progressClosed.Percentage != 100 {
		t.Errorf("expected progress after closing sub-task 1/1 (100%%), got %v", progressClosed)
	}

	// 9. GET /chat/groups/:id/channels/:ticketChannel.Id -> Verify ticket list page
	respViewTickets, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?status=open", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("GET tickets list page failed: %v", err)
	}
	defer respViewTickets.Close()

	if respViewTickets.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for ticket list page, got %d", respViewTickets.StatusCode())
	}
	if !strings.Contains(respViewTickets.String(), "Fix login button glitch") {
		t.Errorf("expected ticket list page to contain open issue title 'Fix login button glitch', got: %s", respViewTickets.String())
	}

	// 10. GET /chat/groups/:id/channels/:issue1.Id -> Verify issue detail page
	respViewDetail, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("GET issue detail page failed: %v", err)
	}
	defer respViewDetail.Close()

	if respViewDetail.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for issue detail page, got %d", respViewDetail.StatusCode())
	}
	if !strings.Contains(respViewDetail.String(), "Update CSS media queries") {
		t.Errorf("expected issue detail page to contain sub-task title 'Update CSS media queries', got: %s", respViewDetail.String())
	}
	if !strings.Contains(respViewDetail.String(), "closed (completed)") {
		t.Errorf("expected issue detail page to contain closing reason 'closed (completed)' for sub-task, got: %s", respViewDetail.String())
	}

	// 11. Edit Ticket 1 title and content with parent_ids
	respEdit1, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":      "Fix login button glitch (Updated Title)",
			"content":    "Updated content for login button issue.",
			"parent_ids": fmt.Sprintf("%d", ticketChannel.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("edit ticket 1 failed: %v", err)
	}
	respEdit1.Close()

	issue1Edited, _ := model.NewChannelFromId(issue1.Id)
	if issue1Edited.Moniker != "Fix login button glitch (Updated Title)" {
		t.Errorf("expected moniker 'Fix login button glitch (Updated Title)', got '%s'", issue1Edited.Moniker)
	}
	if issue1Edited.Description != "Updated content for login button issue." {
		t.Errorf("expected description 'Updated content for login button issue.', got '%s'", issue1Edited.Description)
	}

	// 12. Verify editing ticket without passing parent_ids rejects with alert/422
	respZeroParents, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"title": "Update CSS media queries (Zero Parents)",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", group.Id, subTask.Id))
	if err != nil {
		t.Fatalf("zero parents edit failed: %v", err)
	}
	defer respZeroParents.Close()

	if !strings.Contains(respZeroParents.String(), "The moniker format is invalid.") && respZeroParents.StatusCode() != http.StatusUnprocessableEntity {
		t.Errorf("expected alert message or status 422 when editing without parent_ids, got status %d, body: %s", respZeroParents.StatusCode(), respZeroParents.String())
	}

	// 13. Test Multi-parent association
	respMultiParent, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"parent_ids": fmt.Sprintf("%d", ticketChannel.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", group.Id, subTask.Id))
	if err != nil {
		t.Fatalf("set multi-parent failed: %v", err)
	}
	respMultiParent.Close()

	parents := subTask.ParentChannels()
	if len(parents) != 1 || parents[0].Id != ticketChannel.Id {
		t.Errorf("expected subTask to have 1 parent (%d), got %v", ticketChannel.Id, parents)
	}

	// 14. Test Live Search route /groups/:id/tickets/search
	respSearch, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/tickets/search?q=Fix", group.Id))
	if err != nil {
		t.Fatalf("GET ticket search failed: %v", err)
	}
	defer respSearch.Close()

	if respSearch.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for ticket search, got %d", respSearch.StatusCode())
	}
}
