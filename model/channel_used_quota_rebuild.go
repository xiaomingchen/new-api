package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const historicalFallbackModelRatio = 37.5
const channelUsedQuotaRepairOptionKey = "ChannelUsedQuotaHistoricalFallbackRepaired"

type ChannelUsedQuotaRebuildResult struct {
	ChannelCount        int   `json:"channel_count"`
	UpdatedChannelCount int   `json:"updated_channel_count"`
	IgnoredLogCount     int64 `json:"ignored_log_count"`
	PreviousUsedQuota   int64 `json:"previous_used_quota"`
	RebuiltUsedQuota    int64 `json:"rebuilt_used_quota"`
}

type channelUsedQuotaLogRow struct {
	Id        int    `gorm:"column:id"`
	ChannelId int    `gorm:"column:channel_id"`
	Quota     int    `gorm:"column:quota"`
	Other     string `gorm:"column:other"`
}

type channelUsedQuotaSnapshot struct {
	ModelRatio      float64 `json:"model_ratio"`
	CompletionRatio float64 `json:"completion_ratio"`
}

// RebuildChannelUsedQuota removes the old unconfigured-model fallback charges
// from channel usage totals without changing user balances or consume logs.
func RebuildChannelUsedQuota() (ChannelUsedQuotaRebuildResult, error) {
	result := ChannelUsedQuotaRebuildResult{}
	var completedRepair Option
	if err := DB.Where("key = ?", channelUsedQuotaRepairOptionKey).First(&completedRepair).Error; err == nil {
		return result, errors.New("channel used quota historical fallback repair has already completed")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return result, err
	}

	var channels []Channel
	if err := DB.Select("id, used_quota").Find(&channels).Error; err != nil {
		return result, err
	}
	result.ChannelCount = len(channels)

	quotaByChannel := make(map[int]int64, len(channels))
	var logs []channelUsedQuotaLogRow
	if err := LOG_DB.Model(&Log{}).
		Select("id, channel_id, quota, other").
		Where("type = ? AND channel_id <> 0", LogTypeConsume).
		FindInBatches(&logs, 1000, func(tx *gorm.DB, batch int) error {
			for _, log := range logs {
				var snapshot channelUsedQuotaSnapshot
				if log.Other != "" && common.UnmarshalJsonStr(log.Other, &snapshot) == nil &&
					snapshot.ModelRatio == historicalFallbackModelRatio && snapshot.CompletionRatio == 1 {
					result.IgnoredLogCount++
					quotaByChannel[log.ChannelId] += int64(log.Quota)
					continue
				}
			}
			return nil
		}).Error; err != nil {
		return result, err
	}

	for _, channel := range channels {
		fallbackQuota := quotaByChannel[channel.Id]
		result.PreviousUsedQuota += channel.UsedQuota
		if fallbackQuota > channel.UsedQuota {
			result.RebuiltUsedQuota += 0
		} else {
			result.RebuiltUsedQuota += channel.UsedQuota - fallbackQuota
		}
		if fallbackQuota == 0 {
			continue
		}
		if err := DB.Model(&Channel{}).Where("id = ?", channel.Id).Update("used_quota", gorm.Expr("CASE WHEN used_quota > ? THEN used_quota - ? ELSE 0 END", fallbackQuota, fallbackQuota)).Error; err != nil {
			return result, fmt.Errorf("adjust channel %d used quota: %w", channel.Id, err)
		}
		result.UpdatedChannelCount++
	}
	if err := DB.Create(&Option{Key: channelUsedQuotaRepairOptionKey, Value: "true"}).Error; err != nil {
		return result, err
	}

	return result, nil
}
