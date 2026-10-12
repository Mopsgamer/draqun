package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type ChannelType string

const (
	ChannelTypeText ChannelType = "text"
)

func (ct ChannelType) IsValid() bool {
	return ct == ChannelTypeText
}

type Channel struct {
	checkEmpty
	Id          uint64      `db:"id" goqu:"skipinsert"`
	Name        Name        `db:"name"`
	Moniker     Moniker     `db:"moniker"`
	Description Description `db:"description"`
	Type        ChannelType `db:"type"`
	CreatedAt   TimePast    `db:"created_at"`
	IsDeleted   bool        `db:"is_deleted"`
}

var _ Model = (*Channel)(nil)

func NewChannel(name Name, moniker Moniker, description Description, channelType ChannelType) Channel {
	if channelType == "" {
		channelType = ChannelTypeText
	}
	return Channel{
		Name:        name,
		Moniker:     moniker,
		Description: description,
		Type:        channelType,
		CreatedAt:   TimePast(time.Now()),
		IsDeleted:   false,
	}
}

func NewChannelFromId(id uint64) (Channel, error) {
	channel := Channel{}
	return channel, First(TableChannels, goqu.Ex{"id": id, "is_deleted": false}, &channel)
}

func (channel Channel) Validate() htmx.Alert {
	if !channel.Moniker.IsValid() {
		return htmx.AlertFormatMoniker
	}
	if !channel.Name.IsValid() {
		return htmx.AlertFormatName
	}
	if !channel.Type.IsValid() {
		return htmx.AlertFormatGroupMode
	}
	if !channel.Description.IsValid() {
		return htmx.AlertFormatDescription
	}
	if !channel.CreatedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (channel Channel) IsEmpty() bool {
	return channel.Id == 0
}

func (channel Channel) IsAvailable() bool {
	return channel.Id != 0 && !channel.IsDeleted
}

func (channel *Channel) Insert() error {
	return InsertId(TableChannels, channel, &channel.Id)
}

func (channel Channel) Update() error {
	return Update(TableChannels, channel, goqu.Ex{"id": channel.Id})
}

func (channel Channel) Delete() error {
	channel.IsDeleted = true
	return channel.Update()
}

// Group Channels relation
type GroupChannel struct {
	GroupId   uint64 `db:"group_id"`
	ChannelId uint64 `db:"channel_id"`
	Position  uint   `db:"position"`
}

func AddGroupChannel(groupId, channelId uint64, position uint) error {
	gc := GroupChannel{
		GroupId:   groupId,
		ChannelId: channelId,
		Position:  position,
	}
	return Insert(TableGroupChannels, &gc)
}

func GetGroupChannels(groupId uint64) ([]Channel, error) {
	channels := []Channel{}
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableGroupChannels).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableGroupChannels+".channel_id")))).
		LeftJoin(goqu.I(TableChannelRelations), goqu.On(goqu.I(TableChannelRelations+".child_id").Eq(goqu.I(TableChannels+".id")))).
		Where(goqu.And(
			goqu.I(TableGroupChannels+".group_id").Eq(groupId),
			goqu.I(TableChannels+".is_deleted").Eq(false),
			goqu.I(TableChannelRelations+".child_id").Is(nil),
		)).
		Order(goqu.I(TableGroupChannels + ".position").Asc(), goqu.I(TableChannels + ".id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&channels, sql, args...)
	for i := range channels {
		channels[i].SetFullfilled()
	}
	return channels, err
}

func GetAllGroupChannels(groupId uint64) ([]Channel, error) {
	channels := []Channel{}
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableGroupChannels).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableGroupChannels+".channel_id")))).
		Where(goqu.Ex{
			TableGroupChannels + ".group_id": groupId,
			TableChannels + ".is_deleted":   false,
		}).
		Order(goqu.I(TableGroupChannels + ".position").Asc(), goqu.I(TableChannels + ".id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&channels, sql, args...)
	for i := range channels {
		channels[i].SetFullfilled()
	}
	return channels, err
}

func RemoveGroupChannel(groupId, channelId uint64) error {
	return Delete(TableGroupChannels, goqu.Ex{"group_id": groupId, "channel_id": channelId})
}

// Channel Relations (subchannels / threads)
type ChannelRelation struct {
	ParentId     uint64 `db:"parent_id"`
	ChildId      uint64 `db:"child_id"`
	RelationType string `db:"relation_type"`
}

func AddChannelRelation(parentId, childId uint64, relationType string) error {
	if relationType == "" {
		relationType = "subchannel"
	}
	cr := ChannelRelation{
		ParentId:     parentId,
		ChildId:      childId,
		RelationType: relationType,
	}
	return Insert(TableChannelRelations, &cr)
}

func GetSubchannels(parentId uint64) ([]Channel, error) {
	channels := []Channel{}
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableChannelRelations).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableChannelRelations+".child_id")))).
		Where(goqu.Ex{
			TableChannelRelations + ".parent_id": parentId,
			TableChannels + ".is_deleted":        false,
		}).
		Order(goqu.I(TableChannels + ".id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&channels, sql, args...)
	return channels, err
}

func GetParentChannel(childId uint64) (*Channel, error) {
	channel := new(Channel)
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableChannelRelations).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableChannelRelations+".parent_id")))).
		Where(goqu.Ex{
			TableChannelRelations + ".child_id": childId,
			TableChannels + ".is_deleted":       false,
		}).
		Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(channel)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func RemoveChannelRelation(parentId, childId uint64) error {
	return Delete(TableChannelRelations, goqu.Ex{"parent_id": parentId, "child_id": childId})
}

// User Wall Channel
type UserChannel struct {
	UserId    uint64 `db:"user_id"`
	ChannelId uint64 `db:"channel_id"`
}

func SetUserWallChannel(userId, channelId uint64) error {
	_ = Delete(TableUserChannels, goqu.Ex{"user_id": userId})
	uc := UserChannel{
		UserId:    userId,
		ChannelId: channelId,
	}
	return Insert(TableUserChannels, &uc)
}

func GetUserWallChannel(userId uint64) (*Channel, error) {
	channel := new(Channel)
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableUserChannels).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableUserChannels+".channel_id")))).
		Where(goqu.Ex{
			TableUserChannels + ".user_id": userId,
			TableChannels + ".is_deleted": false,
		}).
		Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(channel)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

// Direct Channel
type DirectChannel struct {
	User1Id   uint64 `db:"user1_id"`
	User2Id   uint64 `db:"user2_id"`
	ChannelId uint64 `db:"channel_id"`
}

func GetOrCreateDirectChannel(user1Id, user2Id uint64) (*Channel, error) {
	if user1Id > user2Id {
		user1Id, user2Id = user2Id, user1Id
	}
	dc := DirectChannel{}
	err := First(TableDirectChannels, goqu.Ex{"user1_id": user1Id, "user2_id": user2Id}, &dc)
	if err == nil && dc.ChannelId != 0 {
		ch, errCh := NewChannelFromId(dc.ChannelId)
		if errCh == nil && ch.IsAvailable() {
			return &ch, nil
		}
	}

	ch := NewChannel("dm", Moniker("Direct Message"), Description("Direct channel"), ChannelTypeText)
	if err := ch.Insert(); err != nil {
		return nil, err
	}
	dcNew := DirectChannel{
		User1Id:   user1Id,
		User2Id:   user2Id,
		ChannelId: ch.Id,
	}
	_ = Insert(TableDirectChannels, &dcNew)
	return &ch, nil
}

// Channel Permission Overrides
type ChannelPermission struct {
	ChannelId       uint64       `db:"channel_id"`
	RoleId          uint64       `db:"role_id"`
	PermMessages    PermMessages `db:"perm_messages"`
	PermRoles       PermSwitch   `db:"perm_roles"`
	PermMembers     PermMembers  `db:"perm_members"`
	PermGroupChange PermSwitch   `db:"perm_group_change"`
	PermAdmin       PermSwitch   `db:"perm_admin"`
	Inherit         bool         `db:"inherit"`
}

func SetChannelPermission(cp ChannelPermission) error {
	_ = Delete(TableChannelPermissions, goqu.Ex{"channel_id": cp.ChannelId, "role_id": cp.RoleId})
	return Insert(TableChannelPermissions, &cp)
}

func GetChannelPermission(channelId, roleId uint64) (*ChannelPermission, error) {
	cp := new(ChannelPermission)
	err := First(TableChannelPermissions, goqu.Ex{"channel_id": channelId, "role_id": roleId}, cp)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func EffectiveRolePermission(channelId uint64, groupRole Role) Role {
	cp, err := GetChannelPermission(channelId, groupRole.Id)
	if err != nil || cp == nil || cp.Inherit {
		return groupRole
	}

	role := groupRole
	if cp.PermMessages != "" {
		role.PermMessages = cp.PermMessages
	}
	if cp.PermRoles != "" {
		role.PermRoles = cp.PermRoles
	}
	if cp.PermMembers != "" {
		role.PermMembers = cp.PermMembers
	}
	if cp.PermGroupChange != "" {
		role.PermGroupChange = cp.PermGroupChange
	}
	if cp.PermAdmin != "" {
		role.PermAdmin = cp.PermAdmin
	}
	return role
}

// Channel Messages Queries
func (channel Channel) MessagesPage(page, limit uint) []Message {
	messageList := []Message{}
	from := (page - 1) * limit

	subquery := Goqu.From(TableMessages).
		Where(goqu.Ex{"channel_id": channel.Id}).
		Order(goqu.I("id").Desc()).
		Limit(limit).Offset(from)
	sql, args, err := Goqu.From(subquery).Order(goqu.I("id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return messageList
	}

	err = Sqlx.Select(&messageList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return messageList
}

func (channel Channel) MessageFirst() *Message {
	message := new(Message)

	sql, args, err := Goqu.From(TableMessages).Where(goqu.C("channel_id").Eq(channel.Id)).Order(goqu.C("id").Asc()).Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(message)
	if err != nil {
		handleErr(err)
	}

	return message
}

func (channel Channel) MessageLast() *Message {
	message := new(Message)

	sql, args, err := Goqu.From(TableMessages).Where(goqu.C("channel_id").Eq(channel.Id)).Order(goqu.C("id").Desc()).Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(message)
	if err != nil {
		handleErr(err)
	}

	return message
}
