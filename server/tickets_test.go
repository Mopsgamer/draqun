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

	// 14. Test Live Search route /groups/:id/tickets/search via GET and POST
	respSearch, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/groups/%d/tickets/search?q=Fix", group.Id))
	if err != nil {
		t.Fatalf("GET ticket search failed: %v", err)
	}
	defer respSearch.Close()

	if respSearch.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for ticket search GET, got %d", respSearch.StatusCode())
	}

	// Test POST form ticket search
	respSearchPost, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"q":    "Fix",
			"mode": "parent",
		}).
		Post(fmt.Sprintf("/groups/%d/tickets/search", group.Id))
	if err != nil {
		t.Fatalf("POST ticket search failed: %v", err)
	}
	defer respSearchPost.Close()

	if respSearchPost.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 for ticket search POST form, got %d", respSearchPost.StatusCode())
	}
	if !strings.Contains(respSearchPost.String(), "Fix login button glitch") {
		t.Errorf("expected POST ticket search response to contain ticket moniker, got: %s", respSearchPost.String())
	}

	// Test non-matching ticket search
	respSearchNoMatchTicket, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		Get(fmt.Sprintf("/groups/%d/tickets/search?q=djskdbsidb", group.Id))
	if err != nil {
		t.Fatalf("non-matching ticket search failed: %v", err)
	}
	defer respSearchNoMatchTicket.Close()

	if !strings.Contains(respSearchNoMatchTicket.String(), "No matching tickets found") {
		t.Errorf("expected 'No matching tickets found' for non-matching query, got: %s", respSearchNoMatchTicket.String())
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

	// Set read permission to disable for @everyone via Permission rule
	readOnlyPerm := model.Permission{
		GroupId:  group.Id,
		Name:     model.PermNameWrite,
		Level:    model.PermLevelDisable,
		Priority: 10,
	}
	_ = readOnlyPerm.Insert()
	_ = readOnlyPerm.AddRole(group.Everyone().Id)

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
	// Grant high priority write permission for ticketChannel to readOnlyUser
	overridePerm := model.Permission{
		GroupId:  group.Id,
		Name:     model.PermNameWrite,
		Level:    model.PermLevelAll,
		Priority: 50,
	}
	_ = overridePerm.Insert()
	_ = overridePerm.AddChannel(ticketChannel.Id)
	_ = overridePerm.AddUser(readOnlyUser.Id)

	// Now read-only user SHOULD be allowed to create a ticket in ticketChannel due to override
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

	// Disable write permission for issueOverride
	blockPerm := model.Permission{
		GroupId:  group.Id,
		Name:     model.PermNameWrite,
		Level:    model.PermLevelDisable,
		Priority: 100,
	}
	_ = blockPerm.Insert()
	_ = blockPerm.AddChannel(issueOverride.Id)
	_ = blockPerm.AddUser(readOnlyUser.Id)

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

func TestNewValidationRules(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Setup Owner and Group
	cookieOwner, owner := createTestUserCookie(t, ts, "Validation Owner", "val_owner", "val_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Validation Group", "val_group")

	// Create custom Role
	role := model.Role{
		GroupId: group.Id,
		Name:    "mod",
		Moniker: "Moderator",
	}
	if err := role.Insert(); err != nil {
		t.Fatalf("failed to insert role: %v", err)
	}

	// 1. Assign role to owner
	respAssign, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"role_id": fmt.Sprintf("%d", role.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/roles", group.Id, owner.Id))
	if err != nil {
		t.Fatalf("assign role failed: %v", err)
	}
	respAssign.Close()

	// 2. Attempting to assign the SAME role again should return AlertUseless
	respAssignDup, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"role_id": fmt.Sprintf("%d", role.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/members/%d/roles", group.Id, owner.Id))
	if err != nil {
		t.Fatalf("duplicate role assign failed: %v", err)
	}
	defer respAssignDup.Close()

	if !strings.Contains(respAssignDup.String(), "This action is useless.") && respAssignDup.StatusCode() != http.StatusUnprocessableEntity {
		t.Errorf("expected AlertUseless/422 for duplicate role assign, got status %d, body: %s", respAssignDup.StatusCode(), respAssignDup.String())
	}

	// 3. Member search for role_id should filter out owner (since owner already has role)
	respSearchRoleMember, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"q":       "Validation",
			"role_id": fmt.Sprintf("%d", role.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/members/search", group.Id))
	if err != nil {
		t.Fatalf("search member for role failed: %v", err)
	}
	defer respSearchRoleMember.Close()

	if strings.Contains(respSearchRoleMember.String(), string(owner.Moniker)) {
		t.Errorf("expected member search to filter out user with existing role, got: %s", respSearchRoleMember.String())
	}

	// Create ticket channel & ticket
	ch := model.NewChannel("issues_val", "Issues Val", "Desc", model.ChannelTypeTicket)
	_ = ch.Insert()
	_ = model.AddGroupChannel(group.Id, ch.Id, 1)

	ticket := model.NewChannel("ticket_val", "Ticket Val", "Desc", model.ChannelTypeTicket)
	_ = ticket.Insert()
	_ = model.SetParentChannels(ticket.Id, []uint64{ch.Id})

	// 4. Ticket parent search for exclude_id=ticket.Id should filter out ch.Id (already assigned parent)
	respParentSearch, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"q":          "Issues",
			"mode":       "parent",
			"exclude_id": fmt.Sprintf("%d", ticket.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/tickets/search", group.Id))
	if err != nil {
		t.Fatalf("parent search failed: %v", err)
	}
	defer respParentSearch.Close()

	if strings.Contains(respParentSearch.String(), "Issues Val") {
		t.Errorf("expected ticket search to filter out already assigned parent channel, got: %s", respParentSearch.String())
	}

	// 5. Attempting to edit ticket with empty parent_ids should fail with error
	respEmptyParents, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"title": "Ticket Val Updated",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", group.Id, ticket.Id))
	if err != nil {
		t.Fatalf("empty parents edit failed: %v", err)
	}
	defer respEmptyParents.Close()

	if !strings.Contains(respEmptyParents.String(), "The moniker format is invalid.") && respEmptyParents.StatusCode() != http.StatusUnprocessableEntity {
		t.Errorf("expected error for empty parent_ids edit, got status %d, body: %s", respEmptyParents.StatusCode(), respEmptyParents.String())
	}
}

func TestTicketChannelHeaderRefactor(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Header Owner", "header_owner", "header_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Header Test Group", "header_test_group")

	// Create Ticket Channel
	ch := model.NewChannel("bugs", "Bug Tracker", "Tracker Description In Sidebar", model.ChannelTypeTicket)
	_ = ch.Insert()
	_ = model.AddGroupChannel(group.Id, ch.Id, 1)

	// Create Root Ticket 1
	root1 := model.NewChannel("root_1", "Root Ticket Alpha", "First root ticket", model.ChannelTypeTicket)
	root1.AuthorId = owner.Id
	root1.Status = "open"
	_ = root1.Insert()
	_ = model.SetParentChannels(root1.Id, []uint64{ch.Id})

	// Create Sub-ticket 1 under Root Ticket 1
	sub1 := model.NewChannel("sub_1", "Sub Ticket Beta", "Child issue", model.ChannelTypeTicket)
	sub1.AuthorId = owner.Id
	sub1.Status = "open"
	_ = sub1.Insert()
	_ = model.SetParentChannels(sub1.Id, []uint64{root1.Id})

	// 1. GET ticket channel page default: should show BOTH Root Ticket Alpha AND Sub Ticket Beta (all tickets by default)
	respDefault, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET ticket channel page failed: %v", err)
	}
	bodyDefault := respDefault.String()
	respDefault.Close()

	if !strings.Contains(bodyDefault, "Root Ticket Alpha") {
		t.Errorf("expected root ticket 'Root Ticket Alpha' in default view, got: %s", bodyDefault)
	}
	if !strings.Contains(bodyDefault, "Sub Ticket Beta") {
		t.Errorf("expected sub ticket 'Sub Ticket Beta' in default view, got: %s", bodyDefault)
	}
	if !strings.Contains(bodyDefault, "Tracker Description In Sidebar") {
		t.Errorf("expected channel description in right sidebar, got: %s", bodyDefault)
	}

	// 2. GET ticket channel page with only_parents=true: should show Root Ticket Alpha, but NOT Sub Ticket Beta
	respParentsOnly, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?only_parents=true", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET ticket channel page with only_parents=true failed: %v", err)
	}
	bodyParentsOnly := respParentsOnly.String()
	respParentsOnly.Close()

	if !strings.Contains(bodyParentsOnly, "Root Ticket Alpha") {
		t.Errorf("expected 'Root Ticket Alpha' in only_parents=true view, got: %s", bodyParentsOnly)
	}
	if strings.Contains(bodyParentsOnly, "Sub Ticket Beta") {
		t.Errorf("expected 'Sub Ticket Beta' to NOT be in only_parents=true view, got: %s", bodyParentsOnly)
	}

	// 3. GET ticket channel search with q=Beta and default (all tickets): should match Sub Ticket Beta
	respSearchAll, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?q=Beta", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET ticket channel search with q=Beta failed: %v", err)
	}
	bodySearchAll := respSearchAll.String()
	respSearchAll.Close()

	if !strings.Contains(bodySearchAll, "Sub Ticket Beta") {
		t.Errorf("expected 'Sub Ticket Beta' when searching with q=Beta, got: %s", bodySearchAll)
	}

	// 4. GET ticket channel search with q=Beta and only_parents=true: should NOT match Sub Ticket Beta
	respSearchRoot, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?only_parents=true&q=Beta", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET ticket channel search with q=Beta and only_parents=true failed: %v", err)
	}
	bodySearchRoot := respSearchRoot.String()
	respSearchRoot.Close()

	if strings.Contains(bodySearchRoot, "Sub Ticket Beta") {
		t.Errorf("expected 'Sub Ticket Beta' to NOT be found in only_parents search for Beta, got: %s", bodySearchRoot)
	}
	if !strings.Contains(bodySearchRoot, "No tickets matching \"Beta\"") {
		t.Errorf("expected 'No tickets matching \"Beta\"' message, got: %s", bodySearchRoot)
	}

	// 5. GET ticket channel with created_by_me=true: should match tickets created by owner
	respCreatedBy, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?created_by_me=true", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET ticket channel created_by_me failed: %v", err)
	}
	bodyCreatedBy := respCreatedBy.String()
	respCreatedBy.Close()

	if !strings.Contains(bodyCreatedBy, "Root Ticket Alpha") {
		t.Errorf("expected 'Root Ticket Alpha' for created_by_me=true, got: %s", bodyCreatedBy)
	}

	// Assign owner to root1
	_ = model.AddChannelAssignee(root1.Id, owner.Id)

	// 6. GET ticket channel with assigned_to_me=true: should match root1
	respAssignedTo, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		Get(fmt.Sprintf("/chat/groups/%d/channels/%d?assigned_to_me=true", group.Id, ch.Id))
	if err != nil {
		t.Fatalf("GET ticket channel assigned_to_me failed: %v", err)
	}
	bodyAssignedTo := respAssignedTo.String()
	respAssignedTo.Close()

	if !strings.Contains(bodyAssignedTo, "Root Ticket Alpha") {
		t.Errorf("expected 'Root Ticket Alpha' for assigned_to_me=true, got: %s", bodyAssignedTo)
	}
}

func TestPrivateSupportTicketAccessControl(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// 1. First user registration automatically creates Draqun group
	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "admin_user", "admin@example.com")
	draqunGroup, err := model.NewGroupFromName("draqun")
	if err != nil || draqunGroup.IsEmpty() {
		t.Fatalf("expected draqun group to be created on first registration: %v", err)
	}

	// 2. Register Support User and assign support role
	cookieSupport, supportUser := createTestUserCookie(t, ts, "Support User", "support_user", "supp@example.com")
	respJoinSupp, _ := ts.Client.R().SetHeader("Cookie", cookieSupport).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinSupp.Close()

	supportRole, errRole := model.NewRoleFromName("support", draqunGroup.Id)
	if errRole != nil || supportRole.IsEmpty() {
		t.Fatalf("expected support role to exist in draqun group")
	}
	ra := model.RoleAssignee{RoleId: supportRole.Id, UserId: supportUser.Id}
	_ = ra.Insert()

	// 3. Register Regular User A and Regular User B
	cookieUserA, userA := createTestUserCookie(t, ts, "Regular User A", "user_a", "usera@example.com")
	respJoinA, _ := ts.Client.R().SetHeader("Cookie", cookieUserA).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinA.Close()

	cookieUserB, userB := createTestUserCookie(t, ts, "Regular User B", "user_b", "userb@example.com")
	respJoinB, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinB.Close()

	// Find private-support channel
	channels, errCh := model.GetGroupChannels(draqunGroup.Id)
	if errCh != nil {
		t.Fatalf("failed to fetch channels: %v", errCh)
	}
	var privSuppCh model.Channel
	for _, ch := range channels {
		if ch.Name == "private-support" {
			privSuppCh = ch
			break
		}
	}
	if privSuppCh.IsEmpty() {
		t.Fatalf("private-support channel not found")
	}

	// 4. User A creates Ticket A
	respCreateA, err := ts.Client.R().
		SetHeader("Cookie", cookieUserA).
		SetFormDataWithMap(map[string]string{
			"title":   "Ticket A Private Issue",
			"content": "Account problem for User A",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("User A create ticket failed: %v", err)
	}
	respCreateA.Close()

	// 5. User B creates Ticket B
	respCreateB, err := ts.Client.R().
		SetHeader("Cookie", cookieUserB).
		SetFormDataWithMap(map[string]string{
			"title":   "Ticket B Private Issue",
			"content": "Billing problem for User B",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("User B create ticket failed: %v", err)
	}
	respCreateB.Close()

	subchs, _ := model.GetSubchannels(privSuppCh.Id)
	var ticketA, ticketB model.Channel
	for _, sub := range subchs {
		if sub.Moniker == "Ticket A Private Issue" {
			ticketA = sub
		}
		if sub.Moniker == "Ticket B Private Issue" {
			ticketB = sub
		}
	}
	if ticketA.IsEmpty() || ticketB.IsEmpty() {
		t.Fatalf("expected both ticket A and ticket B to be created")
	}

	// 6. User A checks
	// User A views private-support channel container -> Should see Ticket A, but NOT Ticket B
	respListA, err := ts.Client.R().SetHeader("Cookie", cookieUserA).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("User A view private support failed: %v", err)
	}
	bodyListA := respListA.String()
	respListA.Close()

	if !strings.Contains(bodyListA, "Ticket A Private Issue") {
		t.Errorf("expected User A to see Ticket A in private support list, got: %s", bodyListA)
	}
	if strings.Contains(bodyListA, "Ticket B Private Issue") {
		t.Errorf("expected User A to NOT see Ticket B in private support list, got: %s", bodyListA)
	}

	// User A tries to view Ticket B directly -> Should be redirected to group URL
	respDetailAOnB, err := ts.Client.R().SetHeader("Cookie", cookieUserA).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, ticketB.Id))
	if err != nil {
		t.Fatalf("User A view Ticket B failed: %v", err)
	}
	defer respDetailAOnB.Close()
	if strings.Contains(respDetailAOnB.String(), "Billing problem for User B") {
		t.Errorf("expected User A to be denied view of Ticket B, but body contained Ticket B content: %s", respDetailAOnB.String())
	}

	// User A searches tickets -> Should see Ticket A, but NOT Ticket B
	respSearchA, err := ts.Client.R().SetHeader("Cookie", cookieUserA).Get(fmt.Sprintf("/groups/%d/tickets/search?q=Ticket", draqunGroup.Id))
	if err != nil {
		t.Fatalf("User A ticket search failed: %v", err)
	}
	bodySearchA := respSearchA.String()
	respSearchA.Close()

	if !strings.Contains(bodySearchA, "Ticket A Private Issue") {
		t.Errorf("expected User A search to find Ticket A, got: %s", bodySearchA)
	}
	if strings.Contains(bodySearchA, "Ticket B Private Issue") {
		t.Errorf("expected User A search to NOT find Ticket B, got: %s", bodySearchA)
	}

	// 7. User B checks
	// User B views private-support channel container -> Should see Ticket B, but NOT Ticket A
	respListB, err := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("User B view private support failed: %v", err)
	}
	bodyListB := respListB.String()
	respListB.Close()

	if !strings.Contains(bodyListB, "Ticket B Private Issue") {
		t.Errorf("expected User B to see Ticket B in private support list, got: %s", bodyListB)
	}
	if strings.Contains(bodyListB, "Ticket A Private Issue") {
		t.Errorf("expected User B to NOT see Ticket A in private support list, got: %s", bodyListB)
	}

	// User B tries to view Ticket A directly -> Should be denied
	respDetailBOnA, err := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, ticketA.Id))
	if err != nil {
		t.Fatalf("User B view Ticket A failed: %v", err)
	}
	defer respDetailBOnA.Close()
	if strings.Contains(respDetailBOnA.String(), "Account problem for User A") {
		t.Errorf("expected User B to be denied view of Ticket A, but body contained Ticket A content: %s", respDetailBOnA.String())
	}

	// 8. Support User checks -> Should see BOTH Ticket A and Ticket B
	respListSupp, err := ts.Client.R().SetHeader("Cookie", cookieSupport).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("Support User view private support failed: %v", err)
	}
	bodyListSupp := respListSupp.String()
	respListSupp.Close()

	if !strings.Contains(bodyListSupp, "Ticket A Private Issue") || !strings.Contains(bodyListSupp, "Ticket B Private Issue") {
		t.Errorf("expected Support User to see both Ticket A and Ticket B in private support list, got: %s", bodyListSupp)
	}

	// 9. Admin User checks -> Should see BOTH Ticket A and Ticket B
	respListAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("Admin User view private support failed: %v", err)
	}
	bodyListAdmin := respListAdmin.String()
	respListAdmin.Close()

	if !strings.Contains(bodyListAdmin, "Ticket A Private Issue") {
		t.Errorf("expected Admin User to see Ticket A in private support list, got: %s", bodyListAdmin)
	}
	if !strings.Contains(bodyListAdmin, "Ticket B Private Issue") {
		t.Errorf("expected Admin User to see Ticket B in private support list, got: %s", bodyListAdmin)
	}

	_ = adminUser
	_ = cookieAdmin
	_ = userA
	_ = userB
}

func TestMoveTicketToOtherTicketChannel(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	cookieOwner, owner := createTestUserCookie(t, ts, "Move Owner", "move_owner", "move_owner@example.com")
	group := createTestGroupDirectly(t, int64(owner.Id), "Move Test Group", "move_test_group")

	cookieOther, otherUser := createTestUserCookie(t, ts, "Other User", "other_user", "other_user@example.com")
	memberOther := model.NewMember(group.Id, otherUser.Id, otherUser.Moniker)
	_ = memberOther.Insert()

	// Create Ticket Container A
	respCreateA, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "tickets_a",
			"moniker":     "Tickets A",
			"description": "Container A",
			"type":        "ticket",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create container A failed: %v", err)
	}
	respCreateA.Close()

	// Create Ticket Container B
	respCreateB, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"name":        "tickets_b",
			"moniker":     "Tickets B",
			"description": "Container B",
			"type":        "ticket",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/create", group.Id))
	if err != nil {
		t.Fatalf("create container B failed: %v", err)
	}
	respCreateB.Close()

	channels, _ := model.GetGroupChannels(group.Id)
	var containerA, containerB model.Channel
	for _, ch := range channels {
		if ch.Name == "tickets_a" {
			containerA = ch
		}
		if ch.Name == "tickets_b" {
			containerB = ch
		}
	}
	if containerA.IsEmpty() || containerB.IsEmpty() {
		t.Fatalf("expected containers A and B to be created")
	}

	// Create a ticket under container A
	respCreateTicket, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"title":   "Issue To Move",
			"content": "This ticket will be moved",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", group.Id, containerA.Id))
	if err != nil {
		t.Fatalf("create ticket under A failed: %v", err)
	}
	respCreateTicket.Close()

	subsA, _ := model.GetSubchannels(containerA.Id)
	if len(subsA) != 1 {
		t.Fatalf("expected 1 ticket under container A, got %d", len(subsA))
	}
	ticket := subsA[0]

	// 1. Test search with mode=move: should show container B, but NOT container A (current parent) or the ticket itself
	respSearchMove, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"mode":       "move",
			"exclude_id": fmt.Sprintf("%d", ticket.Id),
			"q":          "Tickets",
		}).
		Post(fmt.Sprintf("/groups/%d/tickets/search", group.Id))
	if err != nil {
		t.Fatalf("ticket move search failed: %v", err)
	}
	defer respSearchMove.Close()

	bodySearch := respSearchMove.String()
	if !strings.Contains(bodySearch, "Tickets B") {
		t.Errorf("expected search in mode=move to find Tickets B, got: %s", bodySearch)
	}
	if strings.Contains(bodySearch, "Tickets A") {
		t.Errorf("expected search in mode=move to filter out Tickets A (current parent), got: %s", bodySearch)
	}
	if strings.Contains(bodySearch, "Issue To Move") {
		t.Errorf("expected search in mode=move to filter out ticket itself, got: %s", bodySearch)
	}

	// 2. Disable write perm for everyone, verify member without write perm gets 403 Forbidden
	noWritePerm := model.Permission{
		GroupId:  group.Id,
		Name:     model.PermNameWrite,
		Level:    model.PermLevelDisable,
		Priority: 10,
	}
	_ = noWritePerm.Insert()
	_ = noWritePerm.AddRole(group.Everyone().Id)

	respUnauthorizedMove, err := ts.Client.R().
		SetHeader("Cookie", cookieOther).
		SetFormDataWithMap(map[string]string{
			"target_channel_id": fmt.Sprintf("%d", containerB.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/move", group.Id, ticket.Id))
	if err != nil {
		t.Fatalf("unauthorized move request failed: %v", err)
	}
	if respUnauthorizedMove.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for unauthorized move, got %d", respUnauthorizedMove.StatusCode())
	}
	respUnauthorizedMove.Close()

	// 3. Move ticket into itself or invalid channel -> expect error
	respInvalidMove, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetFormDataWithMap(map[string]string{
			"target_channel_id": fmt.Sprintf("%d", ticket.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/move", group.Id, ticket.Id))
	if err != nil {
		t.Fatalf("invalid move request failed: %v", err)
	}
	if respInvalidMove.StatusCode() == http.StatusOK {
		t.Errorf("expected failure when moving ticket to itself, got %d", respInvalidMove.StatusCode())
	}
	respInvalidMove.Close()

	// 4. Authorized owner moves ticket to container B
	respMove, err := ts.Client.R().
		SetHeader("Cookie", cookieOwner).
		SetHeader("HX-Request", "true").
		SetFormDataWithMap(map[string]string{
			"target_channel_id": fmt.Sprintf("%d", containerB.Id),
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/move", group.Id, ticket.Id))
	if err != nil {
		t.Fatalf("authorized move ticket failed: %v", err)
	}
	defer respMove.Close()

	if respMove.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK with HX-Redirect for move ticket, got %d", respMove.StatusCode())
	}
	if hxRedirect := respMove.Header("HX-Redirect"); !strings.Contains(hxRedirect, fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, ticket.Id)) {
		t.Errorf("expected HX-Redirect to ticket detail page, got '%s'", hxRedirect)
	}

	// 5. Verify parent relations in DB
	parentsAfterMove, err := model.GetParentChannels(ticket.Id)
	if err != nil {
		t.Fatalf("GetParentChannels failed: %v", err)
	}
	if len(parentsAfterMove) != 1 || parentsAfterMove[0].Id != containerB.Id {
		t.Errorf("expected ticket's parent to be container B (%d), got: %v", containerB.Id, parentsAfterMove)
	}

	subsB, _ := model.GetSubchannels(containerB.Id)
	if len(subsB) != 1 || subsB[0].Id != ticket.Id {
		t.Errorf("expected container B to have 1 subchannel (ticket %d), got: %v", ticket.Id, subsB)
	}

	subsAAfter, _ := model.GetSubchannels(containerA.Id)
	if len(subsAAfter) != 0 {
		t.Errorf("expected container A to have 0 subchannels after move, got: %v", subsAAfter)
	}
}

func TestInternalChannelAccessAndSidebarVisibility(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// 1. First user registration automatically creates Draqun group
	cookieAdmin, _ := createTestUserCookie(t, ts, "Admin Staff", "admin_staff", "admin_staff@example.com")
	draqunGroup, err := model.NewGroupFromName("draqun")
	if err != nil || draqunGroup.IsEmpty() {
		t.Fatalf("expected draqun group to be created on first registration: %v", err)
	}

	// 2. Register Support User and assign support role
	cookieSupport, supportUser := createTestUserCookie(t, ts, "Support Staff", "supp_staff", "supp_staff@example.com")
	respJoinSupp, _ := ts.Client.R().SetHeader("Cookie", cookieSupport).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinSupp.Close()

	supportRole, errRole := model.NewRoleFromName("support", draqunGroup.Id)
	if errRole != nil || supportRole.IsEmpty() {
		t.Fatalf("expected support role to exist in draqun group")
	}
	ra := model.RoleAssignee{RoleId: supportRole.Id, UserId: supportUser.Id}
	_ = ra.Insert()

	// 3. Register Regular User
	cookieRegular, _ := createTestUserCookie(t, ts, "Ordinary Member", "ordinary_member", "ordinary_member@example.com")
	respJoinReg, _ := ts.Client.R().SetHeader("Cookie", cookieRegular).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinReg.Close()

	// Find internal channel
	channels, errCh := model.GetGroupChannels(draqunGroup.Id)
	if errCh != nil {
		t.Fatalf("failed to fetch channels: %v", errCh)
	}
	var internalCh model.Channel
	for _, ch := range channels {
		if ch.Name == "internal" {
			internalCh = ch
			break
		}
	}
	if internalCh.IsEmpty() {
		t.Fatalf("internal channel not found")
	}

	internalUrlPath := fmt.Sprintf("/chat/groups/%d/channels/%d", draqunGroup.Id, internalCh.Id)

	// 4. Regular User checks:
	// a) Viewing group main page -> Should NOT see internal channel link in sidebar
	respRegGroup, err := ts.Client.R().SetHeader("Cookie", cookieRegular).Get(fmt.Sprintf("/chat/groups/%d", draqunGroup.Id))
	if err != nil {
		t.Fatalf("regular user group view failed: %v", err)
	}
	bodyRegGroup := respRegGroup.String()
	respRegGroup.Close()

	if strings.Contains(bodyRegGroup, internalUrlPath) || strings.Contains(bodyRegGroup, "#internal") {
		t.Errorf("expected ordinary user to NOT see internal channel link in sidebar, got: %s", bodyRegGroup)
	}

	// b) Direct navigation to internal channel -> Should be denied (redirected)
	respRegDirect, err := ts.Client.R().SetHeader("Cookie", cookieRegular).Get(internalUrlPath)
	if err != nil {
		t.Fatalf("regular user direct internal view failed: %v", err)
	}
	defer respRegDirect.Close()
	if strings.Contains(respRegDirect.String(), "Private internal discussion channel for moderators") {
		t.Errorf("expected ordinary user to be denied access to internal channel, but content was returned")
	}

	// 5. Support User checks:
	// a) Viewing group main page -> SHOULD see internal channel link in sidebar
	respSuppGroup, err := ts.Client.R().SetHeader("Cookie", cookieSupport).Get(fmt.Sprintf("/chat/groups/%d", draqunGroup.Id))
	if err != nil {
		t.Fatalf("support user group view failed: %v", err)
	}
	bodySuppGroup := respSuppGroup.String()
	respSuppGroup.Close()

	if !strings.Contains(bodySuppGroup, internalUrlPath) {
		t.Errorf("expected support user to see Internal channel link in sidebar, got: %s", bodySuppGroup)
	}

	// b) Direct navigation to internal channel -> Allowed
	respSuppDirect, err := ts.Client.R().SetHeader("Cookie", cookieSupport).Get(internalUrlPath)
	if err != nil {
		t.Fatalf("support user direct internal view failed: %v", err)
	}
	defer respSuppDirect.Close()
	if !strings.Contains(respSuppDirect.String(), "Private internal discussion channel for moderators") {
		t.Errorf("expected support user to have access to internal channel")
	}

	// 6. Admin User checks:
	// Viewing group main page -> SHOULD see internal channel link in sidebar
	respAdminGroup, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/chat/groups/%d", draqunGroup.Id))
	if err != nil {
		t.Fatalf("admin user group view failed: %v", err)
	}
	bodyAdminGroup := respAdminGroup.String()
	respAdminGroup.Close()

	if !strings.Contains(bodyAdminGroup, internalUrlPath) {
		t.Errorf("expected admin user to see Internal channel link in sidebar, got: %s", bodyAdminGroup)
	}
}

func TestAPIAccessControlForPrivateChannelsAndTickets(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// 1. Admin creates Draqun group
	cookieAdmin, _ := createTestUserCookie(t, ts, "API Admin User", "api_admin", "api_admin@example.com")
	draqunGroup, err := model.NewGroupFromName("draqun")
	if err != nil || draqunGroup.IsEmpty() {
		t.Fatalf("expected draqun group to be created on first registration: %v", err)
	}

	// 2. Register Regular User A and Regular User B
	cookieUserA, userA := createTestUserCookie(t, ts, "API User A", "api_user_a", "api_usera@example.com")
	respJoinA, _ := ts.Client.R().SetHeader("Cookie", cookieUserA).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinA.Close()

	cookieUserB, userB := createTestUserCookie(t, ts, "API User B", "api_user_b", "api_userb@example.com")
	respJoinB, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).Put(fmt.Sprintf("/groups/%d/join", draqunGroup.Id))
	respJoinB.Close()

	// Find internal and private-support channels
	channels, errCh := model.GetGroupChannels(draqunGroup.Id)
	if errCh != nil {
		t.Fatalf("failed to fetch channels: %v", errCh)
	}
	var internalCh, privSuppCh model.Channel
	for _, ch := range channels {
		if ch.Name == "internal" {
			internalCh = ch
		}
		if ch.Name == "private-support" {
			privSuppCh = ch
		}
	}
	if internalCh.IsEmpty() || privSuppCh.IsEmpty() {
		t.Fatalf("required channels not found")
	}

	// 3. User A creates Ticket A in private-support
	respCreateA, err := ts.Client.R().
		SetHeader("Cookie", cookieUserA).
		SetFormDataWithMap(map[string]string{
			"title":   "User A Confidential Ticket",
			"content": "Secret information of User A",
		}).
		Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/create", draqunGroup.Id, privSuppCh.Id))
	if err != nil {
		t.Fatalf("User A create ticket failed: %v", err)
	}
	respCreateA.Close()

	subchs, _ := model.GetSubchannels(privSuppCh.Id)
	var ticketA model.Channel
	for _, sub := range subchs {
		if sub.Moniker == "User A Confidential Ticket" {
			ticketA = sub
			break
		}
	}
	if ticketA.IsEmpty() {
		t.Fatalf("ticket A not found")
	}

	// 4. User B attempts API calls on internal channel -> Expect 403 Forbidden
	respMsgPageInt, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", draqunGroup.Id, internalCh.Id))
	if respMsgPageInt.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B getting internal messages, got %d", respMsgPageInt.StatusCode())
	}
	respMsgPageInt.Close()

	respMsgPostInt, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).SetFormDataWithMap(map[string]string{"content": "unauthorized msg"}).Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", draqunGroup.Id, internalCh.Id))
	if respMsgPostInt.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B posting to internal channel, got %d", respMsgPostInt.StatusCode())
	}
	respMsgPostInt.Close()

	respMsgPostGroupInt, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).SetFormDataWithMap(map[string]string{"content": "unauthorized msg", "channel_id": fmt.Sprintf("%d", internalCh.Id)}).Post(fmt.Sprintf("/groups/%d/messages/create", draqunGroup.Id))
	if respMsgPostGroupInt.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B posting to group with internal channel_id, got %d", respMsgPostGroupInt.StatusCode())
	}
	respMsgPostGroupInt.Close()

	// 5. User B attempts API calls on User A's Ticket A -> Expect 403 Forbidden
	respMsgPageTicketA, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", draqunGroup.Id, ticketA.Id))
	if respMsgPageTicketA.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B getting Ticket A messages, got %d", respMsgPageTicketA.StatusCode())
	}
	respMsgPageTicketA.Close()

	respMsgPostTicketA, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).SetFormDataWithMap(map[string]string{"content": "unauthorized comment"}).Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", draqunGroup.Id, ticketA.Id))
	if respMsgPostTicketA.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B posting to Ticket A, got %d", respMsgPostTicketA.StatusCode())
	}
	respMsgPostTicketA.Close()

	respMsgPostGroupTicketA, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).SetFormDataWithMap(map[string]string{"content": "unauthorized comment", "channel_id": fmt.Sprintf("%d", ticketA.Id)}).Post(fmt.Sprintf("/groups/%d/messages/create", draqunGroup.Id))
	if respMsgPostGroupTicketA.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B posting to group with Ticket A channel_id, got %d", respMsgPostGroupTicketA.StatusCode())
	}
	respMsgPostGroupTicketA.Close()

	respEditTicketA, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).SetFormDataWithMap(map[string]string{"title": "Hacked Title", "parent_ids": fmt.Sprintf("%d", privSuppCh.Id)}).Post(fmt.Sprintf("/groups/%d/channels/%d/tickets/edit", draqunGroup.Id, ticketA.Id))
	if respEditTicketA.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B editing Ticket A, got %d", respEditTicketA.StatusCode())
	}
	respEditTicketA.Close()

	respStatusTicketA, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).SetFormDataWithMap(map[string]string{"status": "closed", "closed_reason": "completed"}).Post(fmt.Sprintf("/groups/%d/channels/%d/status", draqunGroup.Id, ticketA.Id))
	if respStatusTicketA.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B changing status of Ticket A, got %d", respStatusTicketA.StatusCode())
	}
	respStatusTicketA.Close()

	// 6. User A (author) performs API calls on Ticket A -> Expect 200 OK
	respMsgPostUserA, _ := ts.Client.R().SetHeader("Cookie", cookieUserA).SetFormDataWithMap(map[string]string{"content": "User A legitimate comment"}).Post(fmt.Sprintf("/groups/%d/channels/%d/messages/create", draqunGroup.Id, ticketA.Id))
	if respMsgPostUserA.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for User A posting to Ticket A, got %d", respMsgPostUserA.StatusCode())
	}
	respMsgPostUserA.Close()

	respMsgPageUserA, _ := ts.Client.R().SetHeader("Cookie", cookieUserA).Get(fmt.Sprintf("/groups/%d/channels/%d/messages/page/1", draqunGroup.Id, ticketA.Id))
	if respMsgPageUserA.StatusCode() != http.StatusOK {
		t.Errorf("expected 200 OK for User A getting Ticket A messages, got %d", respMsgPageUserA.StatusCode())
	}
	respMsgPageUserA.Close()

	// 7. Member search with private channel_id by User B -> Expect 403 Forbidden
	respMemSearchInt, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/groups/%d/members/search?channel_id=%d&q=API", draqunGroup.Id, internalCh.Id))
	if respMemSearchInt.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B member search with internal channel_id, got %d", respMemSearchInt.StatusCode())
	}
	respMemSearchInt.Close()

	respMemSearchTicketA, _ := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/groups/%d/members/search?channel_id=%d&q=API", draqunGroup.Id, ticketA.Id))
	if respMemSearchTicketA.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User B member search with Ticket A channel_id, got %d", respMemSearchTicketA.StatusCode())
	}
	respMemSearchTicketA.Close()

	_ = cookieAdmin
	_ = userA
	_ = userB
}
