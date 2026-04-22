package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvancedCustomValidateResponsesToChatConverterPath(t *testing.T) {
	valid := &AdvancedCustomConfig{
		Routes: []AdvancedCustomRoute{
			{
				IncomingPath: "/v1/responses",
				UpstreamPath: "/v1/chat/completions",
				Converter:    AdvancedCustomConverterOpenAIResponsesToOpenAIChatCompletions,
			},
		},
	}
	require.NoError(t, valid.Validate())

	tests := []struct {
		name         string
		incomingPath string
	}{
		{name: "chat completions", incomingPath: "/v1/chat/completions"},
		{name: "responses compact", incomingPath: "/v1/responses/compact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AdvancedCustomConfig{
				Routes: []AdvancedCustomRoute{
					{
						IncomingPath: tt.incomingPath,
						UpstreamPath: "/v1/chat/completions",
						Converter:    AdvancedCustomConverterOpenAIResponsesToOpenAIChatCompletions,
					},
				},
			}
			err := config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "converter does not match incoming_path")
		})
	}
}

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
