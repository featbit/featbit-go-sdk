package datasynchronization

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNextDelay(t *testing.T) {
	strategy := NewWithFirstRetryDelay(time.Second)
	strategy.SetGoodRunAtNow()
	delay := strategy.NextDelay()
	assert.True(t, delay < time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 2*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 4*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 8*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 16*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 32*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 60*time.Second)
}
