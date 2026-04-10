package model

import "time"

type channelTokenUsageRow struct {
	ChannelID       int   `gorm:"column:channel_id"`
	UsedTokens      int64 `gorm:"column:used_tokens"`
	UsedTokensToday int64 `gorm:"column:used_tokens_today"`
}

func PopulateChannelTokenUsage(channels []*Channel) error {
	if len(channels) == 0 {
		return nil
	}

	channelIDs := make([]int, 0, len(channels))
	for _, channel := range channels {
		if channel == nil || channel.Id == 0 {
			continue
		}
		channel.UsedTokens = 0
		channel.UsedTokensToday = 0
		channelIDs = append(channelIDs, channel.Id)
	}
	if len(channelIDs) == 0 {
		return nil
	}

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	var rows []channelTokenUsageRow
	err := LOG_DB.Model(&Log{}).
		Select(
			"channel_id, "+
				"COALESCE(SUM(prompt_tokens), 0) + COALESCE(SUM(completion_tokens), 0) AS used_tokens, "+
				"COALESCE(SUM(CASE WHEN created_at >= ? THEN prompt_tokens ELSE 0 END), 0) + "+
				"COALESCE(SUM(CASE WHEN created_at >= ? THEN completion_tokens ELSE 0 END), 0) AS used_tokens_today",
			startOfToday,
			startOfToday,
		).
		Where("type = ?", LogTypeConsume).
		Where("channel_id IN ?", channelIDs).
		Group("channel_id").
		Scan(&rows).Error
	if err != nil {
		return err
	}

	usageByChannelID := make(map[int]channelTokenUsageRow, len(rows))
	for _, row := range rows {
		usageByChannelID[row.ChannelID] = row
	}

	for _, channel := range channels {
		if channel == nil {
			continue
		}
		if row, ok := usageByChannelID[channel.Id]; ok {
			channel.UsedTokens = row.UsedTokens
			channel.UsedTokensToday = row.UsedTokensToday
		}
	}

	return nil
}
