package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUsers(t *testing.T) {
	s := State{
		users: []string{"hello", "world"},
	}
	assert.Equal(t, "*hello world", s.get_users(), "Get users failed")
}
