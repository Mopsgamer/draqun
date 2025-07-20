package database

import (
	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Create new DB record.
func (db Database) GroupCreate(group model_database.Group) *uint64 {
	return Insert(db, TableGroups, group)
}

// Change the existing DB record.
func (db Database) GroupUpdate(group model_database.Group) bool {
	return Update(db, TableGroups, group, exp.Ex{"id": group.Id})
}

// Delete the existing DB record and memberships.
func (db Database) GroupDelete(groupId uint64) bool {
	deleted := false
	deleted = Delete(db, TableGroups, goqu.Ex{"id": groupId})
	if !deleted {
		return deleted
	}

	deleted = Delete(db, TableMembers, goqu.Ex{"group_id": groupId})
	if !deleted {
		return deleted
	}

	// TODO: delete roles(!) and messages(?) when deleting group

	return true
}

// Get the group by her identificator.
func (db Database) GroupById(groupId uint64) *model_database.Group {
	return First[model_database.Group](db, TableGroups, goqu.Ex{"id": groupId})
}

// Get the group by her group name.
func (db Database) GroupByName(groupName string) *model_database.Group {
	return First[model_database.Group](db, TableGroups, goqu.Ex{"name": groupName})
}
