package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/require"
)

func TestChannelSettingsGetProxyURL(t *testing.T) {
	oldSetting := append([]system_setting.ProxyPoolItem(nil), system_setting.GetProxyPoolSetting().Proxies...)
	t.Cleanup(func() {
		system_setting.GetProxyPoolSetting().Proxies = oldSetting
	})

	system_setting.GetProxyPoolSetting().Proxies = []system_setting.ProxyPoolItem{
		{Id: "pool-a", Name: "Pool A", ProxyURL: "socks5://127.0.0.1:7890"},
	}

	require.Equal(t, "socks5://127.0.0.1:7890", ChannelSettings{
		ProxyMode:   ChannelProxyModePool,
		ProxyPoolId: "pool-a",
	}.GetProxyURL())

	require.Equal(t, "http://manual:7890", ChannelSettings{
		Proxy:     "http://manual:7890",
		ProxyMode: ChannelProxyModeCustom,
	}.GetProxyURL())

	require.Equal(t, "http://legacy:7890", ChannelSettings{
		Proxy: "http://legacy:7890",
	}.GetProxyURL())

	require.Equal(t, "", ChannelSettings{
		Proxy:     "http://legacy:7890",
		ProxyMode: ChannelProxyModeNone,
	}.GetProxyURL())
}
