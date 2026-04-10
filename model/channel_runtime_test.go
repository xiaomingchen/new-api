package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPopulateChannelRuntimeStatsFillsCurrentConnectionsAndLastUsedAt(t *testing.T) {
	resetChannelRuntimeStats()
	t.Cleanup(resetChannelRuntimeStats)

	channels := []*Channel{
		{Id: 1},
		{Id: 2},
		{Id: 3},
	}

	before := time.Now().Unix()
	release1 := ObserveChannelRequest(1)
	release2 := ObserveChannelRequest(1)
	release3 := ObserveChannelRequest(2)

	PopulateChannelRuntimeStats(channels)

	assert.EqualValues(t, 2, channels[0].CurrentConnections)
	assert.EqualValues(t, 1, channels[1].CurrentConnections)
	assert.Zero(t, channels[2].CurrentConnections)
	assert.GreaterOrEqual(t, channels[0].LastUsedAt, before)
	assert.GreaterOrEqual(t, channels[1].LastUsedAt, before)
	assert.Zero(t, channels[2].LastUsedAt)

	release1()
	release2()
	release3()

	PopulateChannelRuntimeStats(channels)

	assert.Zero(t, channels[0].CurrentConnections)
	assert.Zero(t, channels[1].CurrentConnections)
	assert.GreaterOrEqual(t, channels[0].LastUsedAt, before)
	assert.GreaterOrEqual(t, channels[1].LastUsedAt, before)
}
