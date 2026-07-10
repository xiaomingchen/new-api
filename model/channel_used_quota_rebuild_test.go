package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildChannelUsedQuotaIgnoresHistoricalFallbackCharges(t *testing.T) {
	truncateTables(t)

	channels := []*Channel{
		{Id: 1, Name: "affected", Key: "sk-affected", Status: common.ChannelStatusEnabled, UsedQuota: 1000},
		{Id: 2, Name: "unchanged", Key: "sk-unchanged", Status: common.ChannelStatusEnabled, UsedQuota: 200},
		{Id: 3, Name: "empty", Key: "sk-empty", Status: common.ChannelStatusEnabled, UsedQuota: 300},
	}
	require.NoError(t, DB.Create(channels).Error)

	logs := []*Log{
		{Type: LogTypeConsume, ChannelId: 1, Quota: 100, Other: "{\"model_ratio\":1.25,\"completion_ratio\":4}"},
		{Type: LogTypeConsume, ChannelId: 1, Quota: 900, Other: "{\"model_ratio\":37.5,\"completion_ratio\":1}"},
		{Type: LogTypeConsume, ChannelId: 2, Quota: 200, Other: "{\"model_ratio\":37.5}"},
		{Type: LogTypeError, ChannelId: 1, Quota: 500, Other: "{\"model_ratio\":37.5,\"completion_ratio\":1}"},
		{Type: LogTypeConsume, ChannelId: 0, Quota: 700, Other: "{\"model_ratio\":1.25,\"completion_ratio\":4}"},
	}
	require.NoError(t, LOG_DB.Create(logs).Error)

	result, err := RebuildChannelUsedQuota()
	require.NoError(t, err)
	assert.Equal(t, 3, result.ChannelCount)
	assert.Equal(t, 1, result.UpdatedChannelCount)
	assert.EqualValues(t, 1, result.IgnoredLogCount)
	assert.EqualValues(t, 1500, result.PreviousUsedQuota)
	assert.EqualValues(t, 600, result.RebuiltUsedQuota)

	var rebuilt []*Channel
	require.NoError(t, DB.Order("id").Find(&rebuilt).Error)
	require.Len(t, rebuilt, 3)
	assert.EqualValues(t, 100, rebuilt[0].UsedQuota)
	assert.EqualValues(t, 200, rebuilt[1].UsedQuota)
	assert.EqualValues(t, 300, rebuilt[2].UsedQuota)

	_, err = RebuildChannelUsedQuota()
	require.Error(t, err)
}
