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

func TestMergePreferHigher(t *testing.T) {
	assert := assert.New(t)
	// Higher priority role merged second overrides lower priority role merged first
	role := Role{PermMessages: PermMessagesDelete}
	role.Merge(Role{PermMessages: PermMessagesWrite})
	assert.Equal(role, Role{PermMessages: PermMessagesWrite})

	// Neutral permission inherits lower priority role setting
	roleNeutral := Role{PermMessages: PermMessagesDelete}
	roleNeutral.Merge(Role{PermMessages: PermMessagesNone})
	assert.Equal(roleNeutral, Role{PermMessages: PermMessagesDelete})

	roleSwitch := Role{PermRoles: PermSwitchDisallow}
	roleSwitch.Merge(Role{PermRoles: PermSwitchAllow})
	assert.Equal(roleSwitch, Role{PermRoles: PermSwitchAllow})
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	role := Role{PermMessages: PermMessagesDelete}
	role.Merge(Role{PermMembers: PermMembersDelete})
	assert.Equal(role, Role{PermMessages: PermMessagesDelete, PermMembers: PermMembersDelete})
}
