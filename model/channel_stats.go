package model

import (
	"fmt"
	"sort"
)

type ChannelStats struct {
	ChannelID    int     `json:"channel_id"`
	ChannelName  string  `json:"channel_name"`
	ModelName    string  `json:"model_name"`
	UsedTokens   int64   `json:"used_tokens"`
	RequestCount int64   `json:"request_count"`
	TodayAmount  int64   `json:"today_amount"`
	TotalAmount  int64   `json:"total_amount"`
	SuccessCount int64   `json:"success_count"`
	ErrorCount   int64   `json:"error_count"`
	SuccessRate  float64 `json:"success_rate"`
}

type channelStatsRow struct {
	ChannelID    int    `gorm:"column:channel_id"`
	ModelName    string `gorm:"column:model_name"`
	UsedTokens   int64  `gorm:"column:used_tokens"`
	SuccessCount int64  `gorm:"column:success_count"`
	ErrorCount   int64  `gorm:"column:error_count"`
	RequestCount int64  `gorm:"column:request_count"`
	TodayAmount  int64  `gorm:"column:today_amount"`
}

type channelMetaRow struct {
	ID        int    `gorm:"column:id"`
	Name      string `gorm:"column:name"`
	UsedQuota int64  `gorm:"column:used_quota"`
}

func GetChannelStats(startTimestamp int64, endTimestamp int64) ([]*ChannelStats, error) {
	rows := make([]channelStatsRow, 0)
	query := LOG_DB.Table("logs").
		Select(
			`channel_id,
			model_name,
			SUM(CASE WHEN type = ? THEN prompt_tokens + completion_tokens ELSE 0 END) AS used_tokens,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS error_count,
			SUM(CASE WHEN type IN (?, ?) THEN 1 ELSE 0 END) AS request_count,
			SUM(CASE WHEN type = ? THEN quota ELSE 0 END) AS today_amount`,
			LogTypeConsume,
			LogTypeConsume,
			LogTypeError,
			LogTypeConsume,
			LogTypeError,
			LogTypeConsume,
		).
		Where("channel_id <> 0").
		Where("type IN ?", []int{LogTypeConsume, LogTypeError})

	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		query = query.Where("created_at <= ?", endTimestamp)
	}

	if err := query.Group("channel_id, model_name").Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []*ChannelStats{}, nil
	}

	channelIDSet := make(map[int]struct{}, len(rows))
	channelIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		if _, ok := channelIDSet[row.ChannelID]; ok {
			continue
		}
		channelIDSet[row.ChannelID] = struct{}{}
		channelIDs = append(channelIDs, row.ChannelID)
	}

	channelMeta := make(map[int]channelMetaRow, len(channelIDs))
	metaRows := make([]channelMetaRow, 0, len(channelIDs))
	if err := DB.Model(&Channel{}).
		Select("id, name, used_quota").
		Where("id IN ?", channelIDs).
		Find(&metaRows).Error; err != nil {
		return nil, err
	}
	for _, row := range metaRows {
		channelMeta[row.ID] = row
	}

	type channelStatsAggregate struct {
		ChannelStats
		bestModelRequestCount int64
		bestModelUsedTokens   int64
	}

	aggregates := make(map[int]*channelStatsAggregate, len(channelIDs))
	for _, row := range rows {
		agg := aggregates[row.ChannelID]
		if agg == nil {
			agg = &channelStatsAggregate{}
			aggregates[row.ChannelID] = agg
		}

		agg.ChannelID = row.ChannelID
		agg.UsedTokens += row.UsedTokens
		agg.SuccessCount += row.SuccessCount
		agg.ErrorCount += row.ErrorCount
		agg.RequestCount += row.RequestCount
		agg.TodayAmount += row.TodayAmount

		betterModel := false
		switch {
		case row.RequestCount > agg.bestModelRequestCount:
			betterModel = true
		case row.RequestCount == agg.bestModelRequestCount && row.UsedTokens > agg.bestModelUsedTokens:
			betterModel = true
		case row.RequestCount == agg.bestModelRequestCount && row.UsedTokens == agg.bestModelUsedTokens:
			if agg.ModelName == "" {
				betterModel = row.ModelName != ""
			} else if row.ModelName != "" {
				betterModel = row.ModelName < agg.ModelName
			}
		}
		if betterModel {
			agg.ModelName = row.ModelName
			agg.bestModelRequestCount = row.RequestCount
			agg.bestModelUsedTokens = row.UsedTokens
		}
	}

	stats := make([]*ChannelStats, 0, len(aggregates))
	for channelID, agg := range aggregates {
		meta, ok := channelMeta[channelID]
		channelName := fmt.Sprintf("渠道 #%d", channelID)
		totalAmount := int64(0)
		if ok {
			if meta.Name != "" {
				channelName = meta.Name
			}
			totalAmount = meta.UsedQuota
		}

		successRate := 0.0
		if agg.RequestCount > 0 {
			successRate = float64(agg.SuccessCount) * 100 / float64(agg.RequestCount)
		}

		stats = append(stats, &ChannelStats{
			ChannelID:    channelID,
			ChannelName:  channelName,
			ModelName:    agg.ModelName,
			UsedTokens:   agg.UsedTokens,
			RequestCount: agg.RequestCount,
			TodayAmount:  agg.TodayAmount,
			TotalAmount:  totalAmount,
			SuccessCount: agg.SuccessCount,
			ErrorCount:   agg.ErrorCount,
			SuccessRate:  successRate,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].RequestCount == stats[j].RequestCount {
			if stats[i].UsedTokens == stats[j].UsedTokens {
				return stats[i].ChannelID < stats[j].ChannelID
			}
			return stats[i].UsedTokens > stats[j].UsedTokens
		}
		return stats[i].RequestCount > stats[j].RequestCount
	})

	return stats, nil
}
