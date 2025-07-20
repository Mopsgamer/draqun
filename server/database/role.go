package database

import (
	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
)

type Role struct {
	Db *goqu.Database
	*model_database.Role
}

func NewRole(db *goqu.Database) Role {
	return Role{Db: db}
}

func (role *Role) Insert() bool {
	id := Insert(role.Db, TableRoles, role)
	role.Role.Id = uint32(*id)
	return id != nil
}

func (role *Role) Update() bool {
	return Update(role.Db, TableRoles, role, goqu.Ex{"id": role.Id})
}

func (role *Role) Delete(roleId uint64) bool {
	return Delete(role.Db, TableRoles, goqu.Ex{"id": roleId})
}

func (role *Role) FromId(id uint64) bool {
	role.Role = First[model_database.Role](role.Db, TableRoles, goqu.Ex{"id": id})
	return role.Id != 0
}
