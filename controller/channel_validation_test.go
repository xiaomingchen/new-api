package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
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

func stringPtr(value string) *string {
	return &value
}
