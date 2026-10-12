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

	// 15. Test Assigning and Unassigning Members for Tickets
	// Live search members for ticket assignment
	respSearchAssigneeMember, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/members/search?channel_id=%d&q=Ticket", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("GET member search for assignee failed: %v", err)
	}
	defer respSearchAssigneeMember.Close()

	if respSearchAssigneeMember.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for member search, got %d", respSearchAssigneeMember.StatusCode())
	}
	if !strings.Contains(respSearchAssigneeMember.String(), fmt.Sprintf("/groups/%d/channels/%d/assignees", group.Id, issue1.Id)) {
		t.Errorf("expected member search HTML to contain assignees action URL, got: %s", respSearchAssigneeMember.String())
	}

	// Assign owner as assignee to issue1
	respAssign, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"user_id": fmt.Sprintf("%d", owner.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/assignees", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("assign member failed: %v", err)
	}
	respAssign.Close()

	assignees := issue1.Assignees()
	if len(assignees) != 1 || assignees[0].Id != owner.Id {
		t.Errorf("expected 1 assignee with ID %d, got %v", owner.Id, assignees)
	}

	// Verify issue detail page renders assignee moniker
	respViewDetailAssignee, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("GET issue detail page with assignee failed: %v", err)
	}
	defer respViewDetailAssignee.Close()

	if !strings.Contains(respViewDetailAssignee.String(), string(owner.Moniker)) {
		t.Errorf("expected issue detail page to contain assignee moniker '%s', got: %s", owner.Moniker, respViewDetailAssignee.String())
	}

	// Unassign owner from issue1
	respUnassign, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Delete(fmt.Sprintf("/groups/%d/channels/%d/assignees/%d", group.Id, issue1.Id, owner.Id))
	if err != nil {
		t.Fatalf("unassign member failed: %v", err)
	}
	respUnassign.Close()

	assigneesAfter := issue1.Assignees()
	if len(assigneesAfter) != 0 {
		t.Errorf("expected 0 assignees after unassigning, got %d", len(assigneesAfter))
	}
}

func TestTicketPermissionsAndEdgeCases(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Register Owner User and create Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Perm Ticket Owner", "perm_ticket_owner", "perm_ticket_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Perm Ticket Group", "perm_ticket_group")

	// Register Member User (Read-Only)
	cookieReadOnly, readOnlyUser := createTestUserCookie(t, ts, "ReadOnly Member", "readonly_member", "readonly_member@example.com")
	respJoin, err := ts.Client.R().
		SetHeader("Cookie", cookieReadOnly).
		Put(fmt.Sprintf("/groups/%d/join", group.Id))
	if err != nil {
		t.Fatalf("readonly user join group failed: %v", err)
	}
	respJoin.Close()

	// Register Non-Member User
	cookieNonMember, _ := createTestUserCookie(t, ts, "Non Member User", "non_member_user", "non_member@example.com")

	// Set @everyone role to PermMessagesRead (read-only)
	everyoneRole := group.Everyone()
	everyoneRole.PermMessages = model.PermMessagesRead
	if err := everyoneRole.Update(); err != nil {
		t.Fatalf("failed to update everyone role to read-only: %v", err)
	}

	// Create top-level Ticket channel as Owner
	respCreateTicketCh, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "issues_perm",
			"moniker":     "Bug Tracker Perm",
			"description": "Group issue tracker perms",
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

	// 1. Non-member attempts to create a ticket -> 404 Not Found (AlertGroupMemberNotFound)
	respNonMemberCreate, err := ts.Client.R().
		SetHeader("Cookie", cookieNonMember).
		SetFormDataWithMap(map[string]string{
			"title":   "Non member issue",
			"content": "This should fail.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("non member create ticket failed: %v", err)
	}
	if respNonMemberCreate.StatusCode() != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for non-member creating ticket, got %d", respNonMemberCreate.StatusCode())
	}
	respNonMemberCreate.Close()

	// 2. Read-only member attempts to create a ticket -> 403 Forbidden
	respReadOnlyCreate, err := ts.Client.R().
		SetHeader("Cookie", cookieReadOnly).
		SetFormDataWithMap(map[string]string{
			"title":   "Read only member issue",
			"content": "This should fail.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("readonly create ticket failed: %v", err)
	}
	if respReadOnlyCreate.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for read-only member creating ticket, got %d", respReadOnlyCreate.StatusCode())
	}
	respReadOnlyCreate.Close()

	// 3. Create ticket as Owner
	respCreateIssue, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":   "Owner Created Ticket",
			"content": "Valid ticket content.",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("owner create ticket failed: %v", err)
	}
	respCreateIssue.Close()

	issues, _ := model.GetSubchannels(ticketChannel.Id)
	if len(issues) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(issues))
	}
	issue1 := issues[0]

	// 4. Read-only member attempts to edit ticket -> 403 Forbidden
	respReadOnlyEdit, err := ts.Client.R().
		SetHeader("Cookie", cookieReadOnly).
		SetFormDataWithMap(map[string]string{
			"title":      "Attempted edit",
			"parent_ids": fmt.Sprintf("%d", ticketChannel.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("readonly edit ticket failed: %v", err)
	}
	if respReadOnlyEdit.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for read-only member editing ticket, got %d", respReadOnlyEdit.StatusCode())
	}
	respReadOnlyEdit.Close()

	// 5. Read-only member attempts to update ticket status -> 403 Forbidden
	respReadOnlyStatus, err := ts.Client.R().
		SetHeader("Cookie", cookieReadOnly).
		SetFormDataWithMap(map[string]string{
			"status":        "closed",
			"closed_reason": "completed",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/status", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("readonly status change failed: %v", err)
	}
	if respReadOnlyStatus.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for read-only member changing ticket status, got %d", respReadOnlyStatus.StatusCode())
	}
	respReadOnlyStatus.Close()

	// 6. Channel Permission Override Test for Ticket Channel
	// Set ChannelPermission on ticketChannel for @everyone to PermMessagesWrite
	err = model.SetChannelPermission(model.ChannelPermission{
		ChannelId:    ticketChannel.Id,
		RoleId:       everyoneRole.Id,
		PermMessages: model.PermMessagesWrite,
		Inherit:      false,
	})
	if err != nil {
		t.Fatalf("failed to set channel permission override: %v", err)
	}

	// Now read-only user (who has @everyone role) SHOULD be allowed to create a ticket in ticketChannel due to override
	respOverrideCreate, err := ts.Client.R().
		SetHeader("Cookie", cookieReadOnly).
		SetFormDataWithMap(map[string]string{
			"title":   "Allowed via channel override",
			"content": "Channel override worked!",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("override create ticket failed: %v", err)
	}
	if respOverrideCreate.StatusCode() != http.StatusOK && respOverrideCreate.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected 200/303 for user creating ticket with channel write override, got %d", respOverrideCreate.StatusCode())
	}
	respOverrideCreate.Close()

	issues, _ = model.GetSubchannels(ticketChannel.Id)
	if len(issues) != 2 {
		t.Fatalf("expected 2 tickets after override creation, got %d", len(issues))
	}

	var issueOverride model.Channel
	for _, sub := range issues {
		if sub.Moniker == "Allowed via channel override" {
			issueOverride = sub
			break
		}
	}
	if issueOverride.IsEmpty() {
		t.Fatalf("failed to find ticket created via override")
	}

	// Test channel permission override blocking write (PermMessagesRead on issueOverride for @everyone)
	err = model.SetChannelPermission(model.ChannelPermission{
		ChannelId:    issueOverride.Id,
		RoleId:       everyoneRole.Id,
		PermMessages: model.PermMessagesRead,
		Inherit:      false,
	})
	if err != nil {
		t.Fatalf("failed to set channel permission override on issue: %v", err)
	}

	// Read-only member attempts to edit issueOverride -> Should be denied (403 Forbidden)
	respDeniedEditOverride, err := ts.Client.R().
		SetHeader("Cookie", cookieReadOnly).
		SetFormDataWithMap(map[string]string{
			"title":      "Denied edit via channel override",
			"parent_ids": fmt.Sprintf("%d", ticketChannel.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", group.Id, issueOverride.Id))
	if err != nil {
		t.Fatalf("denied edit override failed: %v", err)
	}
	if respDeniedEditOverride.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for member editing ticket with read-only channel override, got %d", respDeniedEditOverride.StatusCode())
	}
	respDeniedEditOverride.Close()

	// 7. Additional Status Reason Test: Close ticket as "not_planned"
	respNotPlanned, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"status":        "closed",
			"closed_reason": "not_planned",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/status", group.Id, issue1.Id))
	if err != nil {
		t.Fatalf("close ticket as not_planned failed: %v", err)
	}
	respNotPlanned.Close()

	issue1NotPlanned, _ := model.NewChannelFromId(issue1.Id)
	if issue1NotPlanned.Status != "closed" || issue1NotPlanned.ClosedReason != "not_planned" {
		t.Errorf("expected issue 1 to be closed as not_planned, got status='%s', reason='%s'", issue1NotPlanned.Status, issue1NotPlanned.ClosedReason)
	}

	// 8. Empty Title Validation on Ticket Creation -> Fails with Alert / FormatMoniker
	respEmptyTitle, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"title":   "   ",
			"content": "Empty title test",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("empty title ticket create failed: %v", err)
	}
	defer respEmptyTitle.Close()

	if !strings.Contains(respEmptyTitle.String(), "The moniker format is invalid.") && respEmptyTitle.StatusCode() != http.StatusUnprocessableEntity {
		t.Errorf("expected error message or 422 for empty title ticket creation, got status %d, body: %s", respEmptyTitle.StatusCode(), respEmptyTitle.String())
	}

	// 9. Ticket Creation on non-existent parent channel -> Alert / 404
	respInvalidParent, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":   "Invalid Parent Ticket",
			"content": "Parent does not exist",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/999999/tickets/create", group.Id))
	if err != nil {
		t.Fatalf("invalid parent create ticket failed: %v", err)
	}
	defer respInvalidParent.Close()

	if respInvalidParent.StatusCode() != http.StatusNotFound && !strings.Contains(respInvalidParent.String(), "not found") {
		t.Errorf("expected 404/not found for ticket creation on non-existent parent, got status %d, body: %s", respInvalidParent.StatusCode(), respInvalidParent.String())
	}

	// 10. Status Filter Tests: GET /chat/groups/:id/channels/:ticketChannel.Id?status=closed and status=all
	respClosedFilter, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?status=closed", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("GET status=closed tickets failed: %v", err)
	}
	defer respClosedFilter.Close()

	if respClosedFilter.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for status=closed filter, got %d", respClosedFilter.StatusCode())
	}
	if !strings.Contains(respClosedFilter.String(), "Owner Created Ticket") {
		t.Errorf("expected closed filter page to contain closed ticket 'Owner Created Ticket', got: %s", respClosedFilter.String())
	}

	respAllFilter, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?status=all", group.Id, ticketChannel.Id))
	if err != nil {
		t.Fatalf("GET status=all tickets failed: %v", err)
	}
	defer respAllFilter.Close()

	if respAllFilter.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for status=all filter, got %d", respAllFilter.StatusCode())
	}

	// 11. Ticket Search with mode and exclude_id options
	respSearchDup, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/tickets/search?mode=duplicate&exclude_id=%d&q=Owner", group.Id, issueOverride.Id))
	if err != nil {
		t.Fatalf("GET duplicate mode ticket search failed: %v", err)
	}
	defer respSearchDup.Close()

	if respSearchDup.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for duplicate ticket search, got %d", respSearchDup.StatusCode())
	}
	if !strings.Contains(respSearchDup.String(), "duplicate_of_id_input") {
		t.Errorf("expected duplicate mode search result to contain duplicate input handler JS, got: %s", respSearchDup.String())
	}

	// Unused user parameter check suppressor
	_ = readOnlyUser
}
