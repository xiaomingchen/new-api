package model

import (
	"fmt"
	"sort"
)

type ChannelStats struct {
	ChannelID    int     `json:"channel_id"`
	ChannelName  string  `json:"channel_name"`
	UsedTokens   int64   `json:"used_tokens"`
	SuccessCount int64   `json:"success_count"`
	ErrorCount   int64   `json:"error_count"`
	RequestCount int64   `json:"request_count"`
	SuccessRate  float64 `json:"success_rate"`
}

type channelStatsRow struct {
	ChannelID    int   `gorm:"column:channel_id"`
	UsedTokens   int64 `gorm:"column:used_tokens"`
	SuccessCount int64 `gorm:"column:success_count"`
	ErrorCount   int64 `gorm:"column:error_count"`
	RequestCount int64 `gorm:"column:request_count"`
}

type channelNameRow struct {
	ID   int    `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

func GetChannelStats(startTimestamp int64, endTimestamp int64) ([]*ChannelStats, error) {
	rows := make([]channelStatsRow, 0)
	query := LOG_DB.Table("logs").
		Select(
			`channel_id,
			SUM(CASE WHEN type = ? THEN prompt_tokens + completion_tokens ELSE 0 END) AS used_tokens,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS error_count,
			SUM(CASE WHEN type IN (?, ?) THEN 1 ELSE 0 END) AS request_count`,
			LogTypeConsume,
			LogTypeConsume,
			LogTypeError,
			LogTypeConsume,
			LogTypeError,
		).
		Where("channel_id <> 0").
		Where("type IN ?", []int{LogTypeConsume, LogTypeError})

	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		query = query.Where("created_at <= ?", endTimestamp)
	}

	if err := query.Group("channel_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []*ChannelStats{}, nil
	}

	channelIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		channelIDs = append(channelIDs, row.ChannelID)
	}

	channelNames := make(map[int]string, len(rows))
	nameRows := make([]channelNameRow, 0, len(rows))
	if err := DB.Model(&Channel{}).
		Select("id, name").
		Where("id IN ?", channelIDs).
		Find(&nameRows).Error; err != nil {
		return nil, err
	}
	for _, row := range nameRows {
		channelNames[row.ID] = row.Name
	}

	stats := make([]*ChannelStats, 0, len(rows))
	for _, row := range rows {
		channelName := channelNames[row.ChannelID]
		if channelName == "" {
			channelName = fmt.Sprintf("渠道 #%d", row.ChannelID)
		}

		successRate := 0.0
		if row.RequestCount > 0 {
			successRate = float64(row.SuccessCount) * 100 / float64(row.RequestCount)
		}

		stats = append(stats, &ChannelStats{
			ChannelID:    row.ChannelID,
			ChannelName:  channelName,
			UsedTokens:   row.UsedTokens,
			SuccessCount: row.SuccessCount,
			ErrorCount:   row.ErrorCount,
			RequestCount: row.RequestCount,
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
