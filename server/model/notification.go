package model

import (
	"fmt"
	"html/template"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/Mopsgamer/draqun/server/session"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

type Notification struct {
	checkEmpty
	Id        uint64   `db:"id" goqu:"skipinsert"`
	UserId    uint64   `db:"user_id"`
	ActorId   uint64   `db:"actor_id"`
	GroupId   uint64   `db:"group_id"`
	ChannelId uint64   `db:"channel_id"`
	MessageId uint64   `db:"message_id"`
	Type      string   `db:"type"`
	Title     string   `db:"title"`
	Content   string   `db:"content"`
	IsRead    bool     `db:"is_read"`
	CreatedAt TimePast `db:"created_at"`
}

var _ Model = (*Notification)(nil)

func NewNotification(userId, actorId, groupId, channelId, messageId uint64, notifType, title, content string) Notification {
	return Notification{
		UserId:    userId,
		ActorId:   actorId,
		GroupId:   groupId,
		ChannelId: channelId,
		MessageId: messageId,
		Type:      notifType,
		Title:     title,
		Content:   content,
		IsRead:    false,
		CreatedAt: TimePast(time.Now()),
	}
}

func (n Notification) Validate() htmx.Alert {
	return nil
}

func (n *Notification) Insert() error {
	return InsertId(TableNotifications, n, &n.Id)
}

func (n Notification) Update() error {
	return Update(TableNotifications, n, goqu.Ex{"id": n.Id})
}

func MarkNotificationRead(id uint64, userId uint64) error {
	return Update(TableNotifications, goqu.Record{"is_read": true}, goqu.Ex{"id": id, "user_id": userId})
}

func MarkAllNotificationsRead(userId uint64) error {
	return Update(TableNotifications, goqu.Record{"is_read": true}, goqu.Ex{"user_id": userId, "is_read": false})
}

func DeleteNotification(id uint64, userId uint64) error {
	return Delete(TableNotifications, goqu.Ex{"id": id, "user_id": userId})
}

func DeleteAllNotifications(userId uint64) error {
	return Delete(TableNotifications, goqu.Ex{"user_id": userId})
}

func GetUnreadNotificationCount(userId uint64) (int, error) {
	sql, args, err := Goqu.Select(goqu.COUNT("*")).From(TableNotifications).Where(goqu.Ex{"user_id": userId, "is_read": false}).ToSQL()
	if err != nil {
		return 0, err
	}
	var count int
	err = Sqlx.Get(&count, sql, args...)
	return count, err
}

func GetUserNotifications(userId uint64, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	sql, args, err := Goqu.From(TableNotifications).Where(goqu.Ex{"user_id": userId}).Order(goqu.C("id").Desc()).Limit(uint(limit)).ToSQL()
	if err != nil {
		return nil, err
	}
	var list []Notification
	err = Sqlx.Select(&list, sql, args...)
	for i := range list {
		list[i].SetFullfilled()
	}
	return list, err
}

func (n Notification) ActorUser() User {
	user, _ := NewUserFromId(n.ActorId)
	return user
}

func (n Notification) TargetGroup() Group {
	group, _ := NewGroupFromId(n.GroupId)
	return group
}

func (n Notification) TargetChannel() Channel {
	channel, _ := NewChannelFromId(n.ChannelId)
	return channel
}

func CreateAndPushNotification(app *fiber.App, notif Notification) error {
	if err := notif.Insert(); err != nil {
		return err
	}

	if app != nil {
		targetUserId := notif.UserId
		unreadCount, _ := GetUnreadNotificationCount(targetUserId)
		badgeOob := render.RenderUnreadBadgeOob(unreadCount)

		targetUser, _ := NewUserFromId(targetUserId)
		notificationsList, _ := GetUserNotifications(targetUserId, 20)
		buf, errRender := render.RenderBuffer(app, "partials/notifications", fiber.Map{
			"Notifications": notificationsList,
			"UnreadCount":   unreadCount,
			"User":          targetUser,
		})

		var listOob string
		if errRender == nil {
			s := buf.String()
			dropdownOob := render.WrapOob("innerHTML:#dropdown-notification-list", &s)
			pageOob := render.WrapOob("innerHTML:#page-notification-list", &s)
			listOob = dropdownOob + pageOob
		}

		targetUrl := fmt.Sprintf("/chat/groups/%d", notif.GroupId)
		if notif.Type == "banned" {
			targetGroup := notif.TargetGroup()
			if !targetGroup.IsEmpty() {
				targetUrl = fmt.Sprintf("/chat/groups/join/%s", targetGroup.Name)
			}
		} else if notif.ChannelId != 0 {
			targetUrl = fmt.Sprintf("/chat/groups/%d/channels/%d", notif.GroupId, notif.ChannelId)
		}

		toastHtml := fmt.Sprintf(`
		<div id="desktop-notification-toast" data-title="%s" data-body="%s" data-url="%s" class="toast toast-end toast-bottom z-50">
			<div class="alert alert-info text-xs shadow-lg flex flex-col items-start gap-1 p-3">
				<div class="font-bold flex items-center gap-1.5"><span class="icon icon-[lucide--bell] size-4"></span>%s</div>
				<div class="line-clamp-2">%s</div>
			</div>
		</div>`, template.HTMLEscapeString(notif.Title), template.HTMLEscapeString(notif.Content), targetUrl, template.HTMLEscapeString(notif.Title), template.HTMLEscapeString(notif.Content))

		toastOob := render.WrapOob("beforeend:body", &toastHtml)

		session.UserSessionMap.PushUser(targetUserId, badgeOob+listOob+toastOob)
	}

	return nil
}
