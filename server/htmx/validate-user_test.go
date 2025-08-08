package htmx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidUserPassword(t *testing.T) {
	assert := assert.New(t)
	assert.Error(IsValidUserPassword(""))
	assert.Error(IsValidUserPassword("1234"))
	assert.NoError(IsValidUserPassword("12341234"))
	assert.NoError(IsValidUserPassword("12341234-{}(){}`"))
}
