package model_database

import "github.com/jmoiron/sqlx/types"

type Right struct {
	Id      uint32 `db:"id"`
	GroupId uint64 `db:"group_id"`

	ChatRead     types.BitBool `db:"perm_chat_read"`
	ChatWrite    types.BitBool `db:"perm_chat_write"`
	ChatDelete   types.BitBool `db:"perm_chat_delete"`
	Kick         types.BitBool `db:"perm_kick"`
	Ban          types.BitBool `db:"perm_ban"`
	ChangeGroup  types.BitBool `db:"perm_change_group"`
	ChangeMember types.BitBool `db:"perm_change_member"`
}

func mergeProp(prop types.BitBool, prop2 types.BitBool) types.BitBool {
	return prop || prop2
}

// Enabled rights have priority over disabled rights.
func (r *Right) Merge(rightsList ...Right) {
	for _, rights := range rightsList {
		r.ChatRead = mergeProp(r.ChatRead, rights.ChatRead)
		r.ChatWrite = mergeProp(r.ChatWrite, rights.ChatWrite)
		r.ChatDelete = mergeProp(r.ChatDelete, rights.ChatDelete)
		r.Kick = mergeProp(r.Kick, rights.Kick)
		r.Ban = mergeProp(r.Ban, rights.Ban)
		r.ChangeGroup = mergeProp(r.ChangeGroup, rights.ChangeGroup)
		r.ChangeMember = mergeProp(r.ChangeMember, rights.ChangeMember)
	}
}
