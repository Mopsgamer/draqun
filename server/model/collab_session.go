package model

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type CollabSessionRecord struct {
	Id             uint64    `db:"id" goqu:"skipinsert"`
	UserId         uint64    `db:"user_id"`
	GroupId        uint64    `db:"group_id"`
	ChannelId      uint64    `db:"channel_id"`
	SubchannelName string    `db:"subchannel_name"`
	InputMuted     bool      `db:"input_muted"`
	OutputMuted    bool      `db:"output_muted"`
	VideoEnabled   bool      `db:"video_enabled"`
	BitrateKbps    int       `db:"bitrate_kbps"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func GetAllCollabSessions() ([]CollabSessionRecord, error) {
	var records []CollabSessionRecord
	sql, args, err := Goqu.From(TableCollabSessions).Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}
	err = Sqlx.Select(&records, sql, args...)
	return records, err
}

func UpsertCollabSession(rec *CollabSessionRecord) error {
	now := time.Now()
	rec.UpdatedAt = now
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}

	var existing CollabSessionRecord
	err := First(TableCollabSessions, goqu.Ex{
		"user_id":    rec.UserId,
		"channel_id": rec.ChannelId,
	}, &existing)

	if err == nil && existing.Id != 0 {
		rec.Id = existing.Id
		return Update(TableCollabSessions, *rec, goqu.Ex{"id": existing.Id})
	}

	return Insert(TableCollabSessions, rec)
}

func DeleteCollabSession(userId, channelId uint64) error {
	return Delete(TableCollabSessions, goqu.Ex{
		"user_id":    userId,
		"channel_id": channelId,
	})
}

func DeleteAllUserCollabSessions(userId uint64) error {
	return Delete(TableCollabSessions, goqu.Ex{
		"user_id": userId,
	})
}
