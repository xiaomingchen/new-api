package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateChannelTokenUsageFillsTotalAndToday(t *testing.T) {
	truncateTables(t)

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	yesterday := startOfToday - 60
	today := startOfToday + 3600

	channels := []*Channel{
		{Id: 1, Name: "channel-1", Key: "sk-1", Status: common.ChannelStatusEnabled},
		{Id: 2, Name: "channel-2", Key: "sk-2", Status: common.ChannelStatusEnabled},
		{Id: 3, Name: "channel-3", Key: "sk-3", Status: common.ChannelStatusEnabled},
	}
	for _, channel := range channels {
		require.NoError(t, DB.Create(channel).Error)
	}

	logs := []*Log{
		{
			UserId:           1,
			CreatedAt:        yesterday,
			Type:             LogTypeConsume,
			ChannelId:        1,
			PromptTokens:     100,
			CompletionTokens: 20,
		},
		{
			UserId:           1,
			CreatedAt:        today,
			Type:             LogTypeConsume,
			ChannelId:        1,
			PromptTokens:     200,
			CompletionTokens: 30,
		},
		{
			UserId:           1,
			CreatedAt:        today,
			Type:             LogTypeError,
			ChannelId:        1,
			PromptTokens:     999,
			CompletionTokens: 999,
		},
		{
			UserId:           1,
			CreatedAt:        today,
			Type:             LogTypeConsume,
			ChannelId:        2,
			PromptTokens:     7,
			CompletionTokens: 3,
		},
	}
	require.NoError(t, DB.Create(logs).Error)

	require.NoError(t, PopulateChannelTokenUsage(channels))

	assert.EqualValues(t, 350, channels[0].UsedTokens)
	assert.EqualValues(t, 230, channels[0].UsedTokensToday)
	assert.EqualValues(t, 10, channels[1].UsedTokens)
	assert.EqualValues(t, 10, channels[1].UsedTokensToday)
	assert.Zero(t, channels[2].UsedTokens)
	assert.Zero(t, channels[2].UsedTokensToday)
}
