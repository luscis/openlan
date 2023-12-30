package cswitch

import (
	"fmt"
	"testing"

	"github.com/luscis/openlan/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestSwitch_LoadPass(t *testing.T) {
	sw := &Switch{}
	cache.User.SetFile("../../.password.no")
	sw.LoadPass()
	cache.User.SetFile("../../packaging/resource/password.example")
	sw.LoadPass()
	for user := range cache.User.List() {
		if user == nil {
			break
		}
		fmt.Printf("%v\n", user)
	}
	assert.Equal(t, 2, cache.User.Users.Len(), "notEqual")
}
