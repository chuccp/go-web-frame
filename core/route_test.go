package core

import (
	"testing"

	"github.com/chuccp/go-web-frame/web"
	"github.com/stretchr/testify/assert"
)

// TODO: web.RouteTree has been removed from the new web package.
// The original test verified RouteTree.Has() which no longer exists.
// This test now verifies basic Route construction compiles.

func TestRoute(t *testing.T) {
	meta := web.NewHandlerMeta()
	assert.NotNil(t, meta)
	assert.False(t, meta.HasAnyKey("key"))

	meta.Add("key", "value")
	assert.True(t, meta.HasAnyKey("key"))
	assert.True(t, meta.HasKeyValue("key", "value"))
	assert.False(t, meta.HasKeyValue("key", "other"))
}
