package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/require"
)

func TestValidateChannelAcceptsProxyWebsiteURL(t *testing.T) {
	channel := &model.Channel{
		Type:       1,
		Key:        "sk-test",
		IsProxy:    true,
		WebsiteURL: stringPtr("https://example.com/channel-config"),
	}

	require.NoError(t, validateChannel(channel, false))
}

func TestValidateChannelRejectsProxyWithoutWebsiteURL(t *testing.T) {
	channel := &model.Channel{
		Type:    1,
		Key:     "sk-test",
		IsProxy: true,
	}

	require.ErrorContains(t, validateChannel(channel, false), "跳转地址")
}

func TestValidateChannelRejectsInvalidWebsiteURL(t *testing.T) {
	channel := &model.Channel{
		Type:       1,
		Key:        "sk-test",
		IsProxy:    true,
		WebsiteURL: stringPtr("example.com/channel-config"),
	}

	require.ErrorContains(t, validateChannel(channel, false), "关联网站地址")
}

func TestValidateChannelProxyPool(t *testing.T) {
	oldSetting := append([]system_setting.ProxyPoolItem(nil), system_setting.GetProxyPoolSetting().Proxies...)
	t.Cleanup(func() {
		system_setting.GetProxyPoolSetting().Proxies = oldSetting
	})

	system_setting.GetProxyPoolSetting().Proxies = []system_setting.ProxyPoolItem{
		{Id: "pool-a", Name: "Pool A", ProxyURL: "http://proxy-a.local:8080"},
	}

	validChannel := &model.Channel{
		Type:    1,
		Key:     "sk-test",
		Setting: stringPtr(`{"proxy_mode":"pool","proxy_pool_id":"pool-a"}`),
	}
	require.NoError(t, validateChannel(validChannel, false))

	missingPoolChannel := &model.Channel{
		Type:    1,
		Key:     "sk-test",
		Setting: stringPtr(`{"proxy_mode":"pool","proxy_pool_id":"missing"}`),
	}
	require.ErrorContains(t, validateChannel(missingPoolChannel, false), "代理池不存在")
}

func stringPtr(value string) *string {
	return &value
}
