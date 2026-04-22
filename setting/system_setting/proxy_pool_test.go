package system_setting

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestApplyProxyPoolSettingJSONAcceptsArray(t *testing.T) {
	old := append([]ProxyPoolItem(nil), GetProxyPoolSetting().Proxies...)
	t.Cleanup(func() {
		GetProxyPoolSetting().Proxies = old
	})

	require.NoError(
		t,
		ApplyProxyPoolSettingJSON(
			`[{"id":"pool-a","name":"Pool A","proxy_url":"http://127.0.0.1:7890"}]`,
		),
	)

	require.Len(t, GetProxyPoolSetting().Proxies, 1)
	require.Equal(t, "pool-a", GetProxyPoolSetting().Proxies[0].Id)
	require.Equal(t, "Pool A", GetProxyPoolSetting().Proxies[0].Name)
	require.Equal(t, "http://127.0.0.1:7890", GetProxyPoolSetting().Proxies[0].ProxyURL)
}

func TestApplyProxyPoolSettingJSONAcceptsLegacyWrappedObject(t *testing.T) {
	old := append([]ProxyPoolItem(nil), GetProxyPoolSetting().Proxies...)
	t.Cleanup(func() {
		GetProxyPoolSetting().Proxies = old
	})

	payload, err := common.Marshal(ProxyPoolSetting{
		Proxies: []ProxyPoolItem{
			{
				Id:       "pool-b",
				Name:     "Pool B",
				ProxyURL: "socks5://127.0.0.1:7891",
			},
		},
	})
	require.NoError(t, err)

	require.NoError(t, ApplyProxyPoolSettingJSON(string(payload)))
	require.Len(t, GetProxyPoolSetting().Proxies, 1)
	require.Equal(t, "pool-b", GetProxyPoolSetting().Proxies[0].Id)
	require.Equal(t, "Pool B", GetProxyPoolSetting().Proxies[0].Name)
	require.Equal(t, "socks5://127.0.0.1:7891", GetProxyPoolSetting().Proxies[0].ProxyURL)
}
