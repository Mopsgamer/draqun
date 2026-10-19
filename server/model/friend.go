package model

import (
	"errors"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
)

type Friendship struct {
	checkEmpty
	Id        uint64   `db:"id" goqu:"skipinsert"`
	UserId    uint64   `db:"user_id"`
	FriendId  uint64   `db:"friend_id"`
	Status    string   `db:"status"`
	CreatedAt TimePast `db:"created_at"`
}

var _ Model = (*Friendship)(nil)

func NewFriendship(userId, friendId uint64, status string) Friendship {
	if status == "" {
		status = FriendStatusPending
	}
	return Friendship{
		UserId:    userId,
		FriendId:  friendId,
		Status:    status,
		CreatedAt: TimePast(time.Now()),
	}
}

func (f Friendship) Validate() htmx.Alert {
	return nil
}

func (f *Friendship) Insert() error {
	return InsertId(TableUserFriends, f, &f.Id)
}

func (f Friendship) Update() error {
	return Update(TableUserFriends, f, goqu.Ex{"id": f.Id})
}

func GetFriendship(userA, userB uint64) (*Friendship, error) {
	f := &Friendship{}
	sql, args, err := Goqu.From(TableUserFriends).Where(
		goqu.Or(
			goqu.Ex{"user_id": userA, "friend_id": userB},
			goqu.Ex{"user_id": userB, "friend_id": userA},
		),
	).ToSQL()
	if err != nil {
		return nil, err
	}
	err = Sqlx.Get(f, sql, args...)
	if err != nil {
		return nil, err
	}
	f.SetFullfilled()
	return f, nil
}

func SendFriendRequest(requesterId, targetId uint64) (*Friendship, error) {
	if requesterId == targetId {
		return nil, errors.New("cannot add yourself as a friend")
	}

	existing, err := GetFriendship(requesterId, targetId)
	if err == nil && existing != nil && !existing.IsEmpty() {
		if existing.Status == FriendStatusAccepted {
			return existing, errors.New("you are already friends with this user")
		}
		if existing.Status == FriendStatusPending {
			if existing.UserId == requesterId {
				return existing, errors.New("friend request already sent")
			}
			existing.Status = FriendStatusAccepted
			if errU := existing.Update(); errU != nil {
				return nil, errU
			}
			return existing, nil
		}
	}

	f := NewFriendship(requesterId, targetId, FriendStatusPending)
	if err := f.Insert(); err != nil {
		return nil, err
	}
	return &f, nil
}

func AcceptFriendRequest(userId, requesterId uint64) error {
	f, err := GetFriendship(userId, requesterId)
	if err != nil || f == nil || f.IsEmpty() {
		return errors.New("friend request not found")
	}
	if f.Status == FriendStatusAccepted {
		return nil
	}
	f.Status = FriendStatusAccepted
	return f.Update()
}

func DeclineOrCancelFriendRequest(userA, userB uint64) error {
	sql, args, err := Goqu.From(TableUserFriends).Delete().Where(
		goqu.Or(
			goqu.Ex{"user_id": userA, "friend_id": userB},
			goqu.Ex{"user_id": userB, "friend_id": userA},
		),
	).ToSQL()
	if err != nil {
		return err
	}
	_, err = Sqlx.Exec(sql, args...)
	return err
}

func RemoveFriend(userA, userB uint64) error {
	return DeclineOrCancelFriendRequest(userA, userB)
}

func GetUserFriends(userId uint64) ([]User, error) {
	sql, args, err := Goqu.From(TableUsers).Where(
		goqu.And(
			goqu.Ex{"is_deleted": false},
			goqu.Or(
				goqu.I(TableUsers+".id").In(
					Goqu.Select("friend_id").From(TableUserFriends).Where(goqu.Ex{"user_id": userId, "status": FriendStatusAccepted}),
				),
				goqu.I(TableUsers+".id").In(
					Goqu.Select("user_id").From(TableUserFriends).Where(goqu.Ex{"friend_id": userId, "status": FriendStatusAccepted}),
				),
			),
		),
	).Order(goqu.I(TableUsers + ".moniker").Asc()).ToSQL()

	if err != nil {
		return nil, err
	}

	var friends []User
	err = Sqlx.Select(&friends, sql, args...)
	for i := range friends {
		friends[i].SetFullfilled()
	}
	return friends, err
}

type FriendRequestInfo struct {
	Friendship Friendship
	Requester  User
}

func GetPendingIncomingRequests(userId uint64) ([]FriendRequestInfo, error) {
	sql, args, err := Goqu.From(TableUserFriends).Where(
		goqu.Ex{
			"friend_id": userId,
			"status":    FriendStatusPending,
		},
	).Order(goqu.I("id").Desc()).ToSQL()

	if err != nil {
		return nil, err
	}

	var friendships []Friendship
	err = Sqlx.Select(&friendships, sql, args...)
	if err != nil {
		return nil, err
	}

	var list []FriendRequestInfo
	for i := range friendships {
		friendships[i].SetFullfilled()
		requester, errU := NewUserFromId(friendships[i].UserId)
		if errU == nil && !requester.IsEmpty() {
			list = append(list, FriendRequestInfo{
				Friendship: friendships[i],
				Requester:  requester,
			})
		}
	}
	return list, nil
}
