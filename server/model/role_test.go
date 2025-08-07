package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeNone(t *testing.T) {
	assert := assert.New(t)
	role := Role{PermMessages: PermMessagesWrite}
	role.Merge(Role{PermMessages: PermMessagesNone})
	assert.Equal(role, Role{PermMessages: PermMessagesWrite})
}

func TestMergePreferWorst(t *testing.T) {
	assert := assert.New(t)
	role := Role{PermMessages: PermMessagesDelete}
	role.Merge(Role{PermMessages: PermMessagesWrite})
	assert.Equal(role, Role{PermMessages: PermMessagesWrite})
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	role := Role{PermMessages: PermMessagesDelete}
	role.Merge(Role{PermMembers: PermMembersDelete})
	assert.Equal(role, Role{PermMessages: PermMessagesDelete, PermMembers: PermMembersDelete})
}
