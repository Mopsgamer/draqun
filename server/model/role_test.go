package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleValidation(t *testing.T) {
	assert := assert.New(t)

	role := Role{
		Name:    "admin",
		Moniker: "Admin",
		Color:   0x00FF00,
	}
	assert.Nil(role.Validate())

	invalidName := Role{
		Name:    "",
		Moniker: "Admin",
	}
	assert.NotNil(invalidName.Validate())
}
