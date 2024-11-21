package model

type Rights struct {
	Id            uint `db:"id"`
	GroupId       uint `db:"group_id"`
	ChatRead      bool `db:"perm_chat_read"`
	ChatWrite     bool `db:"perm_chat_write"`
	ChatDelete    bool `db:"perm_chat_delete"`
	Kick          bool `db:"perm_kick"`
	Ban           bool `db:"perm_ban"`
	ChangeGroup   bool `db:"perm_change_group"`
	ChangeMemeber bool `db:"perm_change_member"`
}
