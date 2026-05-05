package ratelimit

import (
	"testing"

	config2 "github.com/chuccp/go-web-frame/config"
	wf "github.com/chuccp/go-web-frame"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	// Just test that the rate limiter can be added to the builder
	builder := wf.NewBuilder(config2.NewConfig())
	rl := &RateLimit{}
	builder.Service(rl)
	frame := builder.Build()
	assert.NotNil(t, frame)
	// Build succeeds - test passed
}
