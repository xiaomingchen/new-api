package system_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
)

type ProxyPoolItem struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ProxyURL string `json:"proxy_url"`
}

type ProxyPoolSetting struct {
	Proxies []ProxyPoolItem `json:"proxies"`
}

var proxyPoolSetting = ProxyPoolSetting{
	Proxies: []ProxyPoolItem{},
}

func init() {
	config.GlobalConfig.Register("proxy_pool_setting", &proxyPoolSetting)
}

func GetProxyPoolSetting() *ProxyPoolSetting {
	return &proxyPoolSetting
}

func GetProxyPoolByID(id string) (ProxyPoolItem, bool) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return ProxyPoolItem{}, false
	}
	for _, item := range proxyPoolSetting.Proxies {
		if strings.TrimSpace(item.Id) == trimmedID {
			return item, true
		}
	}
	return ProxyPoolItem{}, false
}

func GetProxyPoolURL(id string) string {
	item, ok := GetProxyPoolByID(id)
	if !ok {
		return ""
	}
	return strings.TrimSpace(item.ProxyURL)
}

func NormalizeProxyPoolSetting() {
	NormalizeProxyPoolConfig(&proxyPoolSetting)
}

func NormalizeProxyPoolConfig(setting *ProxyPoolSetting) {
	if setting == nil {
		return
	}

	seenIDs := make(map[string]struct{}, len(setting.Proxies))
	normalized := make([]ProxyPoolItem, 0, len(setting.Proxies))
	for _, item := range setting.Proxies {
		item.Id = strings.TrimSpace(item.Id)
		item.Name = strings.TrimSpace(item.Name)
		item.ProxyURL = strings.TrimSpace(item.ProxyURL)

		if item.Name == "" && item.ProxyURL == "" {
			continue
		}

		if item.Id == "" {
			item.Id = "proxy-" + common.GetUUID()[:12]
		}
		for {
			if _, ok := seenIDs[item.Id]; !ok {
				break
			}
			item.Id = "proxy-" + common.GetUUID()[:12]
		}
		seenIDs[item.Id] = struct{}{}
		normalized = append(normalized, item)
	}
	setting.Proxies = normalized
}
