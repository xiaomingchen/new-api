import { describe, expect, test } from 'bun:test';

import {
  mergeChannelLinkFields,
  resolveChannelLinkFields,
} from './channelLinkFields';

describe('channelLinkFields', () => {
  test('preserves website link for non-proxy channels', () => {
    const result = mergeChannelLinkFields(
      { name: 'demo' },
      {
        is_proxy: false,
        website_url: ' https://example.com/channel ',
      },
    );

    expect(result.is_proxy).toBe(false);
    expect(result.website_url).toBe('https://example.com/channel');
  });

  test('keeps proxy flag and website link together', () => {
    const result = mergeChannelLinkFields(
      {
        is_proxy: true,
        website_url: 'https://example.com/proxy',
      },
      {},
    );

    expect(result.is_proxy).toBe(true);
    expect(result.website_url).toBe('https://example.com/proxy');
  });

  test('does not infer proxy mode from an existing website link', () => {
    const result = resolveChannelLinkFields({
      is_proxy: false,
      website_url: 'https://example.com/site',
    });

    expect(result.is_proxy).toBe(false);
    expect(result.website_url).toBe('https://example.com/site');
  });
});
