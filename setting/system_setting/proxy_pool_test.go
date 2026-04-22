package system_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeProxyPoolConfigGeneratesStableIds(t *testing.T) {
	setting := ProxyPoolSetting{
		Proxies: []ProxyPoolItem{
			{Name: "Proxy A", ProxyURL: "http://proxy-a.local:8080"},
			{Id: "proxy-a", Name: "Proxy B", ProxyURL: "http://proxy-b.local:8080"},
			{Id: "proxy-a", Name: "Proxy C", ProxyURL: "http://proxy-c.local:8080"},
			{Name: "   ", ProxyURL: "   "},
		},
	}

	NormalizeProxyPoolConfig(&setting)

	require.Len(t, setting.Proxies, 3)
	require.NotEmpty(t, setting.Proxies[0].Id)
	require.NotEmpty(t, setting.Proxies[1].Id)
	require.NotEmpty(t, setting.Proxies[2].Id)
	require.NotEqual(t, setting.Proxies[1].Id, setting.Proxies[2].Id)
	require.Equal(t, "Proxy A", setting.Proxies[0].Name)
	require.Equal(t, "http://proxy-a.local:8080", setting.Proxies[0].ProxyURL)
}
