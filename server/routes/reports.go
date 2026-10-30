package routes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

type ReportRequest struct {
	Reason string `json:"reason" form:"reason"`
}

func RouteReports(app *fiber.App) {
	app.Post("/reports/group/:group_id", handleReportGroup)
	app.Post("/reports/user/:user_id", handleReportUser)
	app.Post("/reports/message/:message_id", handleReportMessage)
}

func handleReportGroup(ctx fiber.Ctx) error {
	caller, _ := perms.UserByAuthFromCtx(ctx)
	if caller.IsEmpty() {
		return htmx.AlertUserUnauthorized
	}

	groupID, err := strconv.ParseUint(ctx.Params("group_id"), 10, 64)
	if err != nil || groupID == 0 {
		return htmx.AlertInvalidInput
	}

	targetGroup, errObj := model.NewGroupFromId(groupID)
	if errObj != nil || targetGroup.IsEmpty() {
		return htmx.AlertGroupNotFound
	}

	var req ReportRequest
	if err := ctx.Bind().Body(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		return htmx.AlertInvalidInput
	}

	draqunGroup, errObj := model.GetOrCreateDraqunGroup(caller.Id)
	if errObj != nil {
		return htmx.AlertGroupNotFound
	}

	// Ensure caller is a member of Draqun group
	member := model.NewMember(draqunGroup.Id, caller.Id, model.Moniker(caller.Moniker))
	_ = member.Insert()

	// Find the "reports" ticket channel in Draqun group
	reportsChannel, alert := model.GetOrCreateReportsChannel(draqunGroup.Id)
	if alert != nil || reportsChannel.IsEmpty() {
		return htmx.AlertChannelNotFound
	}

	// Create a ticket subchannel under reportsChannel
	ticketTitle := fmt.Sprintf("Report: Group %s (#%d)", targetGroup.Moniker, targetGroup.Id)
	ticketChannel := model.NewChannel(model.Name("ticket_0"), model.Moniker(ticketTitle), model.Description("Group Report"), model.ChannelTypeTicket)
	ticketChannel.AuthorId = caller.Id
	ticketChannel.Status = "open"

	if err := ticketChannel.Validate(); err != nil {
		return err
	}
	if err := ticketChannel.Insert(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	ticketChannel.Name = model.Name(fmt.Sprintf("ticket_%d", ticketChannel.Id))
	if err := ticketChannel.Update(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	_ = model.SetParentChannels(ticketChannel.Id, []uint64{reportsChannel.Id})

	// Post initial ticket report message
	reportMsgContent := fmt.Sprintf(
		"**[GROUP REPORT]**\n**Reported Group:** %s (@%s, ID: `#%d`)\n**Reporter:** %s (@%s, ID: `#%d`)\n**Reason:** %s",
		targetGroup.Moniker, targetGroup.Name, targetGroup.Id,
		caller.Moniker, caller.Name, caller.Id,
		req.Reason,
	)

	msg := model.NewMessageFilled(ticketChannel.Id, draqunGroup.Id, caller.Id, reportMsgContent)
	if err := msg.Insert(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	return ctx.Render("partials/alert", fiber.Map{
		"Variant": "success",
		"Message": "Group report submitted successfully. Punisher team notified.",
	})
}

func handleReportUser(ctx fiber.Ctx) error {
	caller, _ := perms.UserByAuthFromCtx(ctx)
	if caller.IsEmpty() {
		return htmx.AlertUserUnauthorized
	}

	targetUserID, err := strconv.ParseUint(ctx.Params("user_id"), 10, 64)
	if err != nil || targetUserID == 0 {
		return htmx.AlertInvalidInput
	}

	targetUser, errObj := model.NewUserFromId(targetUserID)
	if errObj != nil || targetUser.IsEmpty() {
		return htmx.AlertUserNotFound
	}

	var req ReportRequest
	if err := ctx.Bind().Body(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		return htmx.AlertInvalidInput
	}

	draqunGroup, errObj := model.GetOrCreateDraqunGroup(caller.Id)
	if errObj != nil {
		return htmx.AlertGroupNotFound
	}

	member := model.NewMember(draqunGroup.Id, caller.Id, model.Moniker(caller.Moniker))
	_ = member.Insert()

	reportsChannel, alert := model.GetOrCreateReportsChannel(draqunGroup.Id)
	if alert != nil || reportsChannel.IsEmpty() {
		return htmx.AlertChannelNotFound
	}

	ticketTitle := fmt.Sprintf("Report: User %s (#%d)", targetUser.Moniker, targetUser.Id)
	ticketChannel := model.NewChannel(model.Name("ticket_0"), model.Moniker(ticketTitle), model.Description("User Report"), model.ChannelTypeTicket)
	ticketChannel.AuthorId = caller.Id
	ticketChannel.Status = "open"

	if err := ticketChannel.Validate(); err != nil {
		return err
	}
	if err := ticketChannel.Insert(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	ticketChannel.Name = model.Name(fmt.Sprintf("ticket_%d", ticketChannel.Id))
	if err := ticketChannel.Update(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	_ = model.SetParentChannels(ticketChannel.Id, []uint64{reportsChannel.Id})

	reportMsgContent := fmt.Sprintf(
		"**[USER REPORT]**\n**Reported User:** %s (@%s, ID: `#%d`)\n**Reporter:** %s (@%s, ID: `#%d`)\n**Reason:** %s",
		targetUser.Moniker, targetUser.Name, targetUser.Id,
		caller.Moniker, caller.Name, caller.Id,
		req.Reason,
	)

	msg := model.NewMessageFilled(ticketChannel.Id, draqunGroup.Id, caller.Id, reportMsgContent)
	if err := msg.Insert(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	return ctx.Render("partials/alert", fiber.Map{
		"Variant": "success",
		"Message": "User report submitted successfully. Punisher team notified.",
	})
}

func handleReportMessage(ctx fiber.Ctx) error {
	caller, _ := perms.UserByAuthFromCtx(ctx)
	if caller.IsEmpty() {
		return htmx.AlertUserUnauthorized
	}

	messageID, err := strconv.ParseUint(ctx.Params("message_id"), 10, 64)
	if err != nil || messageID == 0 {
		return htmx.AlertInvalidInput
	}

	targetMessage, errObj := model.NewMessageFromId(messageID)
	if errObj != nil || targetMessage.IsEmpty() {
		return htmx.AlertMessageNotFound
	}

	var req ReportRequest
	if err := ctx.Bind().Body(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		return htmx.AlertInvalidInput
	}

	draqunGroup, errObj := model.GetOrCreateDraqunGroup(caller.Id)
	if errObj != nil {
		return htmx.AlertGroupNotFound
	}

	member := model.NewMember(draqunGroup.Id, caller.Id, model.Moniker(caller.Moniker))
	_ = member.Insert()

	reportsChannel, alert := model.GetOrCreateReportsChannel(draqunGroup.Id)
	if alert != nil || reportsChannel.IsEmpty() {
		return htmx.AlertChannelNotFound
	}

	author := targetMessage.Author()
	ticketTitle := fmt.Sprintf("Report: Message #%d by %s", targetMessage.Id, author.Moniker)
	ticketChannel := model.NewChannel(model.Name("ticket_0"), model.Moniker(ticketTitle), model.Description("Message Report"), model.ChannelTypeTicket)
	ticketChannel.AuthorId = caller.Id
	ticketChannel.Status = "open"

	if err := ticketChannel.Validate(); err != nil {
		return err
	}
	if err := ticketChannel.Insert(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	ticketChannel.Name = model.Name(fmt.Sprintf("ticket_%d", ticketChannel.Id))
	if err := ticketChannel.Update(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	_ = model.SetParentChannels(ticketChannel.Id, []uint64{reportsChannel.Id})

	reportMsgContent := fmt.Sprintf(
		"**[MESSAGE REPORT]**\n**Reported Message ID:** `#%d`\n**Author:** %s (@%s, ID: `#%d`)\n**Channel ID:** `#%d`\n**Content:**\n> %s\n\n**Reporter:** %s (@%s, ID: `#%d`)\n**Reason:** %s",
		targetMessage.Id,
		author.Moniker, author.Name, author.Id,
		targetMessage.ChannelId,
		string(targetMessage.Content),
		caller.Moniker, caller.Name, caller.Id,
		req.Reason,
	)

	msg := model.NewMessageFilled(ticketChannel.Id, draqunGroup.Id, caller.Id, reportMsgContent)
	if err := msg.Insert(); err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	return ctx.Render("partials/alert", fiber.Map{
		"Variant": "success",
		"Message": "Message report submitted successfully. Punisher team notified.",
	})
}
