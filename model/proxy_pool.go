package model

import (
	"strings"

	"github.com/QuantumNous/new-api/relaykit/dto"
)

// CountProxyPoolUsage returns a map of proxy pool ID -> number of channels using that pool.
func CountProxyPoolUsage() (map[string]int, error) {
	counts := make(map[string]int)
	var channels []Channel
	if err := DB.Select("setting").Find(&channels).Error; err != nil {
		return nil, err
	}
	for _, channel := range channels {
		setting := channel.GetSetting()
		if setting.EffectiveProxyMode() != dto.ChannelProxyModePool {
			continue
		}
		proxyPoolID := strings.TrimSpace(setting.ProxyPoolId)
		if proxyPoolID == "" {
			continue
		}
		counts[proxyPoolID]++
	}
	return counts, nil
}