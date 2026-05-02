export const normalizeWebsiteUrl = (websiteUrl) =>
  typeof websiteUrl === 'string' ? websiteUrl.trim() : '';

export const resolveChannelLinkFields = (channel = {}) => ({
  is_proxy: channel.is_proxy === true,
  website_url: normalizeWebsiteUrl(channel.website_url),
});

export const mergeChannelLinkFields = (formValues = {}, inputs = {}) => {
  const nextInputs = { ...formValues };
  const isProxy =
    nextInputs.is_proxy === true || inputs.is_proxy === true;
  const websiteUrl = normalizeWebsiteUrl(
    nextInputs.website_url ?? inputs.website_url,
  );

  nextInputs.is_proxy = isProxy;
  nextInputs.website_url = websiteUrl;

  return nextInputs;
};
