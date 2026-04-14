package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChannelStatsAggregatesTokensRequestsAndSuccessRate(t *testing.T) {
	truncateTables(t)

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	endOfToday := startOfToday + 86399
	yesterday := startOfToday - 60

	channels := []*Channel{
		{Id: 1, Name: "channel-a", Key: "sk-a", Status: common.ChannelStatusEnabled},
		{Id: 2, Name: "channel-b", Key: "sk-b", Status: common.ChannelStatusEnabled},
	}
	for _, channel := range channels {
		require.NoError(t, DB.Create(channel).Error)
	}

	logs := []*Log{
		{
			UserId:           1,
			CreatedAt:        startOfToday + 10,
			Type:             LogTypeConsume,
			ChannelId:        1,
			PromptTokens:     10,
			CompletionTokens: 5,
		},
		{
			UserId:           1,
			CreatedAt:        startOfToday + 20,
			Type:             LogTypeConsume,
			ChannelId:        1,
			PromptTokens:     7,
			CompletionTokens: 3,
		},
		{
			UserId:    1,
			CreatedAt: startOfToday + 30,
			Type:      LogTypeError,
			ChannelId: 1,
		},
		{
			UserId:           1,
			CreatedAt:        startOfToday + 40,
			Type:             LogTypeConsume,
			ChannelId:        2,
			PromptTokens:     20,
			CompletionTokens: 10,
		},
		{
			UserId:    1,
			CreatedAt: startOfToday + 50,
			Type:      LogTypeError,
			ChannelId: 2,
		},
		{
			UserId:    1,
			CreatedAt: startOfToday + 60,
			Type:      LogTypeError,
			ChannelId: 2,
		},
		{
			UserId:           1,
			CreatedAt:        yesterday,
			Type:             LogTypeConsume,
			ChannelId:        1,
			PromptTokens:     999,
			CompletionTokens: 999,
		},
	}
	require.NoError(t, DB.Create(logs).Error)

	stats, err := GetChannelStats(startOfToday, endOfToday)
	require.NoError(t, err)
	require.Len(t, stats, 2)

	assert.Equal(t, 2, stats[0].ChannelID)
	assert.Equal(t, int64(3), stats[0].RequestCount)
	assert.Equal(t, int64(30), stats[0].UsedTokens)

	statsByChannelID := make(map[int]*ChannelStats, len(stats))
	for _, stat := range stats {
		statsByChannelID[stat.ChannelID] = stat
	}

	channelA := statsByChannelID[1]
	require.NotNil(t, channelA)
	assert.Equal(t, "channel-a", channelA.ChannelName)
	assert.EqualValues(t, 25, channelA.UsedTokens)
	assert.EqualValues(t, 2, channelA.SuccessCount)
	assert.EqualValues(t, 1, channelA.ErrorCount)
	assert.EqualValues(t, 3, channelA.RequestCount)
	assert.InDelta(t, 66.6667, channelA.SuccessRate, 0.0001)

	channelB := statsByChannelID[2]
	require.NotNil(t, channelB)
	assert.Equal(t, "channel-b", channelB.ChannelName)
	assert.EqualValues(t, 30, channelB.UsedTokens)
	assert.EqualValues(t, 1, channelB.SuccessCount)
	assert.EqualValues(t, 2, channelB.ErrorCount)
	assert.EqualValues(t, 3, channelB.RequestCount)
	assert.InDelta(t, 33.3333, channelB.SuccessRate, 0.0001)
}
