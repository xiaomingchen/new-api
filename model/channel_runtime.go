package model

import (
	"sync"
	"sync/atomic"
	"time"
)

type channelRuntimeStats struct {
	currentConnections atomic.Int64
	lastUsedAt         atomic.Int64
}

var channelRuntimeStatsMap sync.Map

func getChannelRuntimeStats(channelID int) *channelRuntimeStats {
	if channelID <= 0 {
		return nil
	}
	stats, _ := channelRuntimeStatsMap.LoadOrStore(channelID, &channelRuntimeStats{})
	return stats.(*channelRuntimeStats)
}

func ObserveChannelRequest(channelID int) func() {
	stats := getChannelRuntimeStats(channelID)
	if stats == nil {
		return func() {}
	}

	stats.lastUsedAt.Store(time.Now().Unix())
	stats.currentConnections.Add(1)

	return func() {
		stats.currentConnections.Add(-1)
	}
}

func PopulateChannelRuntimeStats(channels []*Channel) {
	for _, channel := range channels {
		if channel == nil {
			continue
		}
		channel.CurrentConnections = 0
		channel.LastUsedAt = 0
		if channel.Id == 0 {
			continue
		}

		stats, ok := channelRuntimeStatsMap.Load(channel.Id)
		if !ok {
			continue
		}

		channelStats := stats.(*channelRuntimeStats)
		channel.CurrentConnections = channelStats.currentConnections.Load()
		channel.LastUsedAt = channelStats.lastUsedAt.Load()
	}
}

func resetChannelRuntimeStats() {
	channelRuntimeStatsMap = sync.Map{}
}
