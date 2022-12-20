package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Default(t *testing.T) {
	sess1 := Default()
	sess2 := Default()

	assert.NotEqual(t, sess1.ID, sess2.ID)
}
