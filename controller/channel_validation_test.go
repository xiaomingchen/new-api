package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
)

func TestValidateChannelAcceptsWebsiteURL(t *testing.T) {
	channel := &model.Channel{
		Type:       1,
		Key:        "sk-test",
		WebsiteURL: stringPtr("https://example.com/channel-config"),
	}

	require.NoError(t, validateChannel(channel, false))
}

func TestValidateChannelRejectsInvalidWebsiteURL(t *testing.T) {
	channel := &model.Channel{
		Type:       1,
		Key:        "sk-test",
		WebsiteURL: stringPtr("example.com/channel-config"),
	}

	require.ErrorContains(t, validateChannel(channel, false), "关联网站地址")
}

func stringPtr(value string) *string {
	return &value
}
