package internal

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3/client"
)

var userCounter int32

func createTestUserCookie(t testing.TB, ts *TestServer, moniker, name, email string) (string, model.User) {
	count := atomic.AddInt32(&userCounter, 1)
	phone := fmt.Sprintf("1234567%03d", count)

	resp, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"moniker":          moniker,
			"name":             name,
			"email":            email,
			"phone":            phone,
			"password":         "password123",
			"confirm-password": "password123",
		}).
		Post("/account")
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("failed to register user: status %v", resp.StatusCode())
	}

	user, err := model.NewUserFromName(model.Name(name))
	if err != nil {
		t.Fatalf("failed to fetch created user: %v", err)
	}

	for _, cookie := range resp.Cookies() {
		if string(cookie.Key()) == "Authorization" {
			return string(cookie.Key()) + "=" + string(cookie.Value()), user
		}
	}

	t.Fatalf("Authorization cookie not found in response")
	return "", user
}

func createTestGroupDirectly(t testing.TB, creatorId int64, moniker, name string) model.Group {
	group := model.NewGroup(uint64(creatorId), model.Moniker(moniker), model.Name(name), model.GroupModePublic, "", "Group Description", "")
	if err := group.Validate(); err != nil {
		t.Fatalf("failed to validate group: %v", err)
	}
	if err := group.Insert(); err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	member := model.NewMember(group.Id, uint64(creatorId), "")
	if err := member.Insert(); err != nil {
		t.Fatalf("failed to insert group creator as member: %v", err)
	}

	everyone := model.NewRoleEveryone(group.Id)
	if err := everyone.Insert(); err != nil {
		t.Fatalf("failed to insert role everyone: %v", err)
	}

	return group
}

func TestPageHomepage(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	resp, err := ts.Client.Get("/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(strings.ToLower(bodyStr), "homepage") {
		t.Errorf("expected homepage content, got body: %s", bodyStr)
	}
}

func TestPageTerms(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	resp, err := ts.Client.Get("/terms")
	if err != nil {
		t.Fatalf("GET /terms failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(strings.ToLower(bodyStr), "terms") {
		t.Errorf("expected terms content, got body: %s", bodyStr)
	}
}

func TestPagePrivacy(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	resp, err := ts.Client.Get("/privacy")
	if err != nil {
		t.Fatalf("GET /privacy failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(strings.ToLower(bodyStr), "privacy") {
		t.Errorf("expected privacy content, got body: %s", bodyStr)
	}
}

func TestPageAcknowledgements(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	resp, err := ts.Client.Get("/acknowledgements")
	if err != nil {
		t.Fatalf("GET /acknowledgements failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(strings.ToLower(bodyStr), "acknowledgements") {
		t.Errorf("expected acknowledgements content, got body: %s", bodyStr)
	}
}

func TestPageDocs(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	resp, err := ts.Client.Get("/docs")
	if err != nil {
		t.Fatalf("GET /docs failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(strings.ToLower(bodyStr), "docs") && !strings.Contains(strings.ToLower(bodyStr), "swagger") {
		t.Errorf("expected docs or swagger content, got body: %s", bodyStr)
	}
}

func TestPageSettings(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Unauthorized -> should redirect to / (303 See Other)
	respUnauth, err := ts.Client.Get("/settings", client.Config{MaxRedirects: -1})
	if err != nil {
		t.Fatalf("GET /settings failed: %v", err)
	}
	defer respUnauth.Close()

	if respUnauth.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected status 303 (redirect), got %d", respUnauth.StatusCode())
	}

	// Authorized -> should render settings
	cookie, _ := createTestUserCookie(t, ts, "Jane Doe", "jane", "jane@example.com")

	respAuth, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		Get("/settings")
	if err != nil {
		t.Fatalf("GET /settings (auth) failed: %v", err)
	}
	defer respAuth.Close()

	if respAuth.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", respAuth.StatusCode())
	}

	bodyStr := respAuth.String()
	if !strings.Contains(strings.ToLower(bodyStr), "settings") {
		t.Errorf("expected settings content, got body: %s", bodyStr)
	}
}

func TestPageChat(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Unauthorized -> should render chat-login page
	respUnauth, err := ts.Client.Get("/chat")
	if err != nil {
		t.Fatalf("GET /chat failed: %v", err)
	}
	defer respUnauth.Close()

	if respUnauth.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", respUnauth.StatusCode())
	}

	bodyUnauthStr := respUnauth.String()
	if !strings.Contains(strings.ToLower(bodyUnauthStr), "login") && !strings.Contains(strings.ToLower(bodyUnauthStr), "password") {
		t.Errorf("expected chat login/signup page content, got: %s", bodyUnauthStr)
	}

	// Authorized -> should render chat home page
	cookie, _ := createTestUserCookie(t, ts, "Bob Alice", "bob", "bob@example.com")

	respAuth, err := ts.Client.R().
		SetHeader("Cookie", cookie).
		Get("/chat")
	if err != nil {
		t.Fatalf("GET /chat (auth) failed: %v", err)
	}
	defer respAuth.Close()

	if respAuth.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", respAuth.StatusCode())
	}

	bodyStr := respAuth.String()
	if !strings.Contains(strings.ToLower(bodyStr), "logout") && !strings.Contains(strings.ToLower(bodyStr), "settings") {
		t.Errorf("expected authorized chat page content, got: %s", bodyStr)
	}
}

func TestPageGroupJoin(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create test user and group
	_, user := createTestUserCookie(t, ts, "Charlie Green", "charlie", "charlie@example.com")
	_ = createTestGroupDirectly(t, int64(user.Id), "Awesome Group", "awesome_group")

	// Create a second test user who will view the join page
	cookieVisitor, _ := createTestUserCookie(t, ts, "Visitor User", "visitor", "visitor@example.com")

	// GET group join page
	resp, err := ts.Client.R().
		SetHeader("Cookie", cookieVisitor).
		Get("/chat/groups/join/awesome_group")
	if err != nil {
		t.Fatalf("GET /chat/groups/join/awesome_group failed: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	bodyStr := resp.String()
	if !strings.Contains(strings.ToLower(bodyStr), "join") || !strings.Contains(strings.ToLower(bodyStr), "awesome group") {
		t.Errorf("expected join awesome_group content, got body: %s", bodyStr)
	}
}

func TestPageGroup(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// Create test user and group
	cookieMember, userMember := createTestUserCookie(t, ts, "Alice Cooper", "alice", "alice@example.com")
	group := createTestGroupDirectly(t, int64(userMember.Id), "Secrets Group", "secrets_group")

	// Create a second test user who is NOT a member of the group
	cookieNonMember, _ := createTestUserCookie(t, ts, "Non Member", "nonmember", "nonmember@example.com")

	// Member GET group page -> should render 200 OK
	respMember, err := ts.Client.R().
		SetHeader("Cookie", cookieMember).
		Get(fmt.Sprintf("/chat/groups/%d", group.Id))
	if err != nil {
		t.Fatalf("GET /chat/groups/:id failed: %v", err)
	}
	defer respMember.Close()

	if respMember.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200 for member, got %d", respMember.StatusCode())
	}

	bodyMemberStr := respMember.String()
	if !strings.Contains(strings.ToLower(bodyMemberStr), "secrets group") {
		t.Errorf("expected secrets group content, got body: %s", bodyMemberStr)
	}

	// Non-member GET group page -> should redirect to /chat
	respNonMember, err := ts.Client.R().
		SetHeader("Cookie", cookieNonMember).
		SetMaxRedirects(-1).
		Get(fmt.Sprintf("/chat/groups/%d", group.Id))
	if err != nil {
		t.Fatalf("GET /chat/groups/:id for non-member failed: %v", err)
	}
	defer respNonMember.Close()

	if respNonMember.StatusCode() != http.StatusSeeOther {
		t.Errorf("expected status 303 for non-member, got %d", respNonMember.StatusCode())
	}
}
