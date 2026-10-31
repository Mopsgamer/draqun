package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type ChannelType string

const (
	ChannelTypeText   ChannelType = "text"
	ChannelTypeTicket ChannelType = "ticket"
	ChannelTypeOtel   ChannelType = "otel"
	ChannelTypeConfig ChannelType = "config"
	ChannelTypeWall   ChannelType = "wall"

	OtelVirtualChannelID   uint64 = ^uint64(0)
	ConfigVirtualChannelID uint64 = ^uint64(1)
)

func GetOtelVirtualChannel() Channel {
	return Channel{
		Id:          OtelVirtualChannelID,
		Name:        Name("otel"),
		Moniker:     Moniker("OpenTelemetry"),
		Description: Description("OpenTelemetry traces, metrics, and server performance stats for Draqun system administrators."),
		Type:        ChannelTypeOtel,
		Status:      "open",
	}
}

func GetConfigVirtualChannel() Channel {
	return Channel{
		Id:          ConfigVirtualChannelID,
		Name:        Name("config"),
		Moniker:     Moniker("Server Configuration"),
		Description: Description("Global server configuration settings and site-wide management."),
		Type:        ChannelTypeConfig,
		Status:      "open",
	}
}

func (ct ChannelType) IsValid() bool {
	return ct == ChannelTypeText || ct == ChannelTypeTicket || ct == ChannelTypeOtel || ct == ChannelTypeConfig || ct == ChannelTypeWall
}

type Channel struct {
	checkEmpty
	Id            uint64      `db:"id" goqu:"skipinsert"`
	Name          Name        `db:"name"`
	Moniker       Moniker     `db:"moniker"`
	Description   Description `db:"description"`
	Type          ChannelType `db:"type"`
	Status        string      `db:"status"`
	ClosedReason  string      `db:"closed_reason"`
	AuthorId      uint64      `db:"author_id"`
	DuplicateOfId uint64      `db:"duplicate_of_id"`
	UpdatedAt     TimePast    `db:"updated_at"`
	CreatedAt     TimePast    `db:"created_at"`
	IsDeleted     bool        `db:"is_deleted"`
}

var _ Model = (*Channel)(nil)

func NewChannel(name Name, moniker Moniker, description Description, channelType ChannelType) Channel {
	if channelType == "" {
		channelType = ChannelTypeText
	}
	return Channel{
		Name:          name,
		Moniker:       moniker,
		Description:   description,
		Type:          channelType,
		Status:        "open",
		ClosedReason:  "",
		AuthorId:      0,
		DuplicateOfId: 0,
		CreatedAt:     TimePast(time.Now()),
		IsDeleted:     false,
	}
}

func NewChannelFromId(id uint64) (Channel, error) {
	if id == OtelVirtualChannelID {
		return GetOtelVirtualChannel(), nil
	}
	if id == ConfigVirtualChannelID {
		return GetConfigVirtualChannel(), nil
	}
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
	if !channel.IsWall() && !channel.Description.IsValid() {
		return htmx.AlertFormatDescription
	}
	if !channel.CreatedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (channel Channel) IsWall() bool {
	return channel.Type == ChannelTypeWall
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
		Order(goqu.I(TableGroupChannels+".position").Asc(), goqu.I(TableChannels+".id").Asc()).
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
			TableChannels + ".is_deleted":    false,
		}).
		Order(goqu.I(TableGroupChannels+".position").Asc(), goqu.I(TableChannels+".id").Asc()).
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

func UpdateGroupChannelPosition(groupId, channelId uint64, position uint) error {
	return Update(TableGroupChannels, goqu.Record{"position": position}, goqu.Ex{"group_id": groupId, "channel_id": channelId})
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

func GetChannelGroup(channelId uint64) (*Group, error) {
	group := new(Group)
	sql, args, err := Goqu.Select(TableGroups+".*").From(TableGroupChannels).
		Join(goqu.I(TableGroups), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableGroupChannels+".group_id")))).
		Where(goqu.Ex{
			TableGroupChannels + ".channel_id": channelId,
			TableGroups + ".is_deleted":        false,
		}).
		Limit(1).
		Prepared(true).ToSQL()
	if err == nil {
		if errQuery := Sqlx.QueryRowx(sql, args...).StructScan(group); errQuery == nil {
			group.SetFullfilled()
			return group, nil
		}
	}

	parent, errP := GetParentChannel(channelId)
	if errP == nil && parent != nil && parent.Id != 0 {
		return GetChannelGroup(parent.Id)
	}

	return nil, fmt.Errorf("group not found for channel %d", channelId)
}

type StarredChannel struct {
	Channel Channel
	Group   Group
}

func GetStarredChannels(userId uint64) ([]StarredChannel, error) {
	channels := []Channel{}
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableUserChannels).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableUserChannels+".channel_id")))).
		Where(goqu.Ex{
			TableUserChannels + ".user_id": userId,
			TableChannels + ".is_deleted":  false,
		}).
		Order(goqu.I(TableChannels + ".id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&channels, sql, args...)
	if err != nil {
		return nil, err
	}

	var res []StarredChannel
	for i := range channels {
		channels[i].SetFullfilled()
		g, errG := GetChannelGroup(channels[i].Id)
		if errG == nil && g != nil && !g.IsEmpty() {
			res = append(res, StarredChannel{
				Channel: channels[i],
				Group:   *g,
			})
		}
	}
	return res, nil
}

func AddStarredChannel(userId, channelId uint64) error {
	uc := UserChannel{
		UserId:    userId,
		ChannelId: channelId,
	}
	return Insert(TableUserChannels, &uc)
}

func RemoveStarredChannel(userId, channelId uint64) error {
	return Delete(TableUserChannels, goqu.Ex{"user_id": userId, "channel_id": channelId})
}

func IsChannelStarred(userId, channelId uint64) bool {
	var count int
	sql, args, err := Goqu.From(TableUserChannels).
		Where(goqu.Ex{"user_id": userId, "channel_id": channelId}).
		Select(goqu.COUNT("*")).
		Prepared(true).ToSQL()
	if err != nil {
		return false
	}
	_ = Sqlx.Get(&count, sql, args...)
	return count > 0
}

func GetParentChannels(childId uint64) ([]Channel, error) {
	channels := []Channel{}
	sql, args, err := Goqu.Select(TableChannels+".*").From(TableChannelRelations).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableChannelRelations+".parent_id")))).
		Where(goqu.Ex{
			TableChannelRelations + ".child_id": childId,
			TableChannels + ".is_deleted":       false,
		}).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&channels, sql, args...)
	return channels, err
}

func SetParentChannels(childId uint64, parentIds []uint64) error {
	var validParentIds []uint64
	seen := make(map[uint64]bool)
	for _, pid := range parentIds {
		if pid != 0 && pid != childId && !seen[pid] {
			seen[pid] = true
			validParentIds = append(validParentIds, pid)
		}
	}
	if len(validParentIds) == 0 {
		return htmx.AlertFormatMoniker
	}

	_ = Delete(TableChannelRelations, goqu.Ex{"child_id": childId})
	for _, pid := range validParentIds {
		_ = AddChannelRelation(pid, childId, "subchannel")
	}
	return nil
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
			TableChannels + ".is_deleted":  false,
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

type DirectChannelItem struct {
	Channel   Channel
	OtherUser User
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

func IsUserInDirectChannel(channelId, userId uint64) bool {
	if channelId == 0 || userId == 0 {
		return false
	}
	var count int
	sql, args, err := Goqu.From(TableDirectChannels).
		Where(
			goqu.Ex{"channel_id": channelId},
			goqu.Or(
				goqu.Ex{"user1_id": userId},
				goqu.Ex{"user2_id": userId},
			),
		).
		Select(goqu.COUNT("*")).
		Prepared(true).ToSQL()
	if err != nil {
		return false
	}
	_ = Sqlx.Get(&count, sql, args...)
	return count > 0
}

func GetDirectChannelUsers(channelId uint64) (User, User, error) {
	dc := DirectChannel{}
	err := First(TableDirectChannels, goqu.Ex{"channel_id": channelId}, &dc)
	if err != nil {
		return User{}, User{}, err
	}
	u1, err1 := NewUserFromId(dc.User1Id)
	u2, err2 := NewUserFromId(dc.User2Id)
	if err1 != nil || err2 != nil || u1.IsEmpty() || u2.IsEmpty() {
		return User{}, User{}, fmt.Errorf("direct channel users not found")
	}
	return u1, u2, nil
}

func GetDirectChannelOtherUser(channelId, userId uint64) (*User, error) {
	dc := DirectChannel{}
	err := First(TableDirectChannels, goqu.Ex{"channel_id": channelId}, &dc)
	if err != nil {
		return nil, err
	}

	var otherId uint64
	if dc.User1Id == userId {
		otherId = dc.User2Id
	} else if dc.User2Id == userId {
		otherId = dc.User1Id
	} else {
		// Fallback for staff inspection: return User1 if caller is neither participant
		otherId = dc.User1Id
	}

	u, errU := NewUserFromId(otherId)
	if errU != nil || u.IsEmpty() {
		return nil, fmt.Errorf("other user not found")
	}
	return &u, nil
}

type UserDirectChannelPair struct {
	Channel DirectChannelItem
	User1   User
	User2   User
}

func GetUserDirectChannelPairs(userId uint64, limit int) ([]UserDirectChannelPair, error) {
	if userId == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	var dcs []DirectChannel
	sql, args, err := Goqu.From(TableDirectChannels).
		Where(goqu.Or(
			goqu.Ex{"user1_id": userId},
			goqu.Ex{"user2_id": userId},
		)).
		Order(goqu.I("channel_id").Desc()).
		Limit(uint(limit)).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&dcs, sql, args...)
	if err != nil {
		return nil, err
	}

	var pairs []UserDirectChannelPair
	for _, dc := range dcs {
		ch, errCh := NewChannelFromId(dc.ChannelId)
		if errCh != nil || !ch.IsAvailable() {
			continue
		}
		u1, u2, errUsers := GetDirectChannelUsers(dc.ChannelId)
		if errUsers != nil {
			continue
		}
		var otherUser User
		if userId == u1.Id {
			otherUser = u2
		} else {
			otherUser = u1
		}
		pairs = append(pairs, UserDirectChannelPair{
			Channel: DirectChannelItem{
				Channel:   ch,
				OtherUser: otherUser,
			},
			User1: u1,
			User2: u2,
		})
	}
	return pairs, nil
}

func GetUserDirectChannels(userId uint64) ([]DirectChannelItem, error) {
	pairs, err := GetUserDirectChannelPairs(userId, 50)
	if err != nil {
		return nil, err
	}
	var items []DirectChannelItem
	for _, p := range pairs {
		items = append(items, p.Channel)
	}
	return items, nil
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

// Channel Assignees
type ChannelAssignee struct {
	checkEmpty
	ChannelId uint64 `db:"channel_id"`
	UserId    uint64 `db:"user_id"`
}

func IsChannelAssignee(channelId, userId uint64) bool {
	ca := ChannelAssignee{}
	err := First(TableChannelAssignees, goqu.Ex{"channel_id": channelId, "user_id": userId}, &ca)
	return err == nil && ca.UserId != 0
}

func AddChannelAssignee(channelId, userId uint64) error {
	ca := ChannelAssignee{
		ChannelId: channelId,
		UserId:    userId,
	}
	return Insert(TableChannelAssignees, &ca)
}

func RemoveChannelAssignee(channelId, userId uint64) error {
	return Delete(TableChannelAssignees, goqu.Ex{"channel_id": channelId, "user_id": userId})
}

func GetChannelAssignees(channelId uint64) ([]User, error) {
	users := []User{}
	sql, args, err := Goqu.Select(TableUsers+".*").From(TableChannelAssignees).
		Join(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableChannelAssignees+".user_id")))).
		Where(goqu.Ex{
			TableChannelAssignees + ".channel_id": channelId,
			TableUsers + ".is_deleted":            false,
		}).
		Order(goqu.I(TableUsers + ".id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&users, sql, args...)
	for i := range users {
		users[i].SetFullfilled()
	}
	return users, err
}

func (channel Channel) Assignees() []User {
	users, err := GetChannelAssignees(channel.Id)
	if err != nil {
		return nil
	}
	return users
}

func (channel Channel) IsTicket() bool {
	return channel.Type == ChannelTypeTicket
}

func (channel Channel) IsOtel() bool {
	return channel.Type == ChannelTypeOtel
}

func (channel Channel) IsConfig() bool {
	return channel.Type == ChannelTypeConfig
}

func (channel Channel) AuthorUser() User {
	if channel.AuthorId == 0 {
		return User{}
	}
	u, err := NewUserFromId(channel.AuthorId)
	if err != nil {
		return User{}
	}
	return u
}

func (channel Channel) DuplicateChannel() *Channel {
	if channel.DuplicateOfId == 0 {
		return nil
	}
	ch, err := NewChannelFromId(channel.DuplicateOfId)
	if err != nil || !ch.IsAvailable() {
		return nil
	}
	return &ch
}

func (channel Channel) ParentChannel() *Channel {
	p, err := GetParentChannel(channel.Id)
	if err != nil || !p.IsAvailable() {
		return nil
	}
	return p
}

func (channel Channel) ParentChannels() []Channel {
	pchs, err := GetParentChannels(channel.Id)
	if err != nil {
		return nil
	}
	return pchs
}

func (channel *Channel) ChangeParentRelation(newParentId uint64) error {
	currParent, _ := GetParentChannel(channel.Id)
	if currParent != nil && currParent.Id != 0 {
		_ = RemoveChannelRelation(currParent.Id, channel.Id)
	}
	if newParentId != 0 {
		return AddChannelRelation(newParentId, channel.Id, "subchannel")
	}
	return nil
}

func SearchTicketChannels(groupId uint64, query string, excludeId uint64, limit int) ([]Channel, error) {
	if limit <= 0 {
		limit = 10
	}
	allTickets, err := GetAllTicketChannelsForGroup(groupId)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	var results []Channel

	for _, ch := range allTickets {
		if excludeId != 0 && ch.Id == excludeId {
			continue
		}
		if queryLower == "" {
			results = append(results, ch)
		} else {
			idStr := fmt.Sprintf("%d", ch.Id)
			if strings.Contains(strings.ToLower(string(ch.Moniker)), queryLower) || strings.Contains(idStr, queryLower) {
				results = append(results, ch)
			}
		}
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func GetAllDescendantTickets(parentId uint64) ([]Channel, error) {
	seen := make(map[uint64]bool)
	seen[parentId] = true
	var descendants []Channel

	var traverse func(id uint64)
	traverse = func(id uint64) {
		subchs, err := GetSubchannels(id)
		if err != nil {
			return
		}
		for _, sub := range subchs {
			if !seen[sub.Id] {
				seen[sub.Id] = true
				if sub.IsTicket() {
					descendants = append(descendants, sub)
				}
				traverse(sub.Id)
			}
		}
	}

	traverse(parentId)
	return descendants, nil
}

func GetAllTicketChannelsForGroup(groupId uint64) ([]Channel, error) {
	groupChs, err := GetAllGroupChannels(groupId)
	if err != nil {
		return nil, err
	}

	seen := make(map[uint64]bool)
	var ticketChs []Channel

	var traverse func(ch Channel)
	traverse = func(ch Channel) {
		if seen[ch.Id] {
			return
		}
		seen[ch.Id] = true
		if ch.IsTicket() {
			ticketChs = append(ticketChs, ch)
		}
		subchs, _ := GetSubchannels(ch.Id)
		for _, sub := range subchs {
			traverse(sub)
		}
	}

	for _, ch := range groupChs {
		traverse(ch)
	}

	return ticketChs, nil
}

func (channel Channel) CommentCount() int {
	var count int
	sql, args, err := Goqu.From(TableMessages).
		Where(goqu.Ex{"channel_id": channel.Id}).
		Select(goqu.COUNT("*")).
		Prepared(true).ToSQL()
	if err != nil {
		return 0
	}
	_ = Sqlx.Get(&count, sql, args...)
	return count
}

type TicketProgress struct {
	ClosedCount int
	TotalCount  int
	Percentage  int
}

func (channel Channel) GetTicketProgress() TicketProgress {
	subchs, err := GetSubchannels(channel.Id)
	if err != nil || len(subchs) == 0 {
		return TicketProgress{ClosedCount: 0, TotalCount: 0, Percentage: 0}
	}

	closed := 0
	total := len(subchs)
	for _, sub := range subchs {
		if sub.Status == "closed" {
			closed++
		}
	}
	percent := (closed * 100) / total
	return TicketProgress{
		ClosedCount: closed,
		TotalCount:  total,
		Percentage:  percent,
	}
}

func GetTicketSubchannels(parentId uint64, statusFilter string) ([]Channel, error) {
	channels := []Channel{}
	whereEx := goqu.Ex{
		TableChannelRelations + ".parent_id": parentId,
		TableChannels + ".is_deleted":        false,
	}
	if statusFilter == "open" {
		whereEx[TableChannels+".status"] = "open"
	} else if statusFilter == "closed" {
		whereEx[TableChannels+".status"] = "closed"
	}

	sql, args, err := Goqu.Select(TableChannels+".*").From(TableChannelRelations).
		Join(goqu.I(TableChannels), goqu.On(goqu.I(TableChannels+".id").Eq(goqu.I(TableChannelRelations+".child_id")))).
		Where(whereEx).
		Order(goqu.I(TableChannels + ".id").Desc()).
		Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = Sqlx.Select(&channels, sql, args...)
	return channels, err
}

func GetOrCreateReportsChannel(groupId uint64) (Channel, htmx.Alert) {
	channels, err := GetGroupChannels(groupId)
	if err == nil {
		for _, ch := range channels {
			if ch.Name == "reports" {
				return ch, nil
			}
		}
	}
	return Channel{}, htmx.AlertChannelNotFound
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
