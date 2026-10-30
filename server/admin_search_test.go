package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Mopsgamer/draqun/server/model"
	"github.com/stretchr/testify/assert"
)

func TestAdminSearchAndAccessBypass(t *testing.T) {
	app := testSetupApp(t)
	ts := startTestServer(t, app)

	// 1. Create global admin user (ID 1)
	cookieAdmin, adminUser := createTestUserCookie(t, ts, "Admin User", "admin_search", "admin_s@example.com")
	assert.Equal(t, uint64(1), adminUser.Id)

	// 2. Create regular user A & regular user B
	_, userA := createTestUserCookie(t, ts, "User A", "usera_s", "usera_s@example.com")
	cookieUserB, userB := createTestUserCookie(t, ts, "User B", "userb_s", "userb_s@example.com")

	// 3. Create private group owned by User A
	grpA := model.NewGroup(
		userA.Id,
		model.Moniker("Secret Group A"),
		model.Name("secretgroupa"),
		model.GroupModePrivate,
		model.OptionalPasswordHashed(""),
		model.Description("Private secrets"),
		model.Avatar(""),
	)
	errG := grpA.Insert()
	assert.NoError(t, errG)
	m := model.NewMember(grpA.Id, userA.Id, "")
	_ = m.Insert()
	chA := model.NewChannel(model.Name("secretch"), model.Moniker("Secret Channel"), model.Description("secret"), model.ChannelTypeText)
	_ = chA.Insert()
	_ = model.AddGroupChannel(grpA.Id, chA.Id, 1)

	// User A and User B start a direct message channel
	dmCh, errDm := model.GetOrCreateDirectChannel(userA.Id, userB.Id)
	assert.NoError(t, errDm)
	dmMsg := model.NewMessageFilled(dmCh.Id, 0, userA.Id, "Top secret direct message")
	_ = dmMsg.Insert()

	// 4. Test non-staff (User B) attempting Admin Search and viewing Group A
	respSearchDenied, err := ts.Client.R().SetHeader("Cookie", cookieUserB).Get("/chat/admin/search")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, respSearchDenied.StatusCode()) // Redirected to /chat

	respGrpB, err := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/chat/groups/%d", grpA.Id))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, respGrpB.StatusCode()) // Redirected away since not member

	respDmB, err := ts.Client.R().SetHeader("Cookie", cookieUserB).Get(fmt.Sprintf("/chat/dm/%d", dmCh.Id))
	assert.NoError(t, err)
	// User B is participant in DM, so B can access it
	assert.Equal(t, http.StatusOK, respDmB.StatusCode())

	// Create User C who is NOT participant in DM
	cookieUserC, userC := createTestUserCookie(t, ts, "User C", "userc_s", "userc_s@example.com")
	_ = userC

	respDmC, err := ts.Client.R().SetHeader("Cookie", cookieUserC).Get(fmt.Sprintf("/chat/dm/%d", dmCh.Id))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, respDmC.StatusCode()) // Redirected away since not participant or staff

	// 5. Test Global Admin (User 1) performing Admin Search and accessing private group & DM
	respSearchAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get("/chat/admin/search?q=Secret")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respSearchAdmin.StatusCode())
	assert.Contains(t, respSearchAdmin.String(), "Secret Group A")

	respGrpAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/chat/groups/%d", grpA.Id))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respGrpAdmin.StatusCode())
	assert.Contains(t, respGrpAdmin.String(), "Secret Group A")

	respDmAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/chat/dm/%d", dmCh.Id))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respDmAdmin.StatusCode())

	respDmMsgAdmin, err := ts.Client.R().SetHeader("Cookie", cookieAdmin).Get(fmt.Sprintf("/groups/dm/channels/%d/messages/page/1", dmCh.Id))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respDmMsgAdmin.StatusCode())
	assert.Contains(t, respDmMsgAdmin.String(), "Top secret direct message")

	// 6. Test Punisher role user in official Draqun group
	draqunGroup, errD := model.GetOrCreateDraqunGroup(adminUser.Id)
	assert.NoError(t, errD)

	punisherRole, errP := model.NewRoleFromName("punisher", draqunGroup.Id)
	assert.NoError(t, errP)

	cookiePun, punUser := createTestUserCookie(t, ts, "Punisher Staff", "pun_staff", "pun_staff@example.com")
	raPun := model.RoleAssignee{
		RoleId: punisherRole.Id,
		UserId: punUser.Id,
	}
	_ = raPun.Insert()

	// Verify punisher staff can access admin search and read DMs
	respSearchPun, err := ts.Client.R().SetHeader("Cookie", cookiePun).Get("/chat/admin/search?q=Secret")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respSearchPun.StatusCode())

	respDmPun, err := ts.Client.R().SetHeader("Cookie", cookiePun).Get(fmt.Sprintf("/chat/dm/%d", dmCh.Id))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respDmPun.StatusCode())
}
