/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../helpers';
import {
  Banner,
  Button,
  Col,
  Input,
  Row,
  Spin,
  Tag,
  Tooltip,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconDelete,
  IconPlus,
  IconRefresh,
  IconSave,
} from '@douyinfe/semi-icons';

const { Text } = Typography;

const QUALITY_COLORS = {
  优秀: 'green',
  良好: 'blue',
  一般: 'orange',
  较差: 'red',
  不可用: 'red',
  未知: 'grey',
};

function createLocalId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `proxy-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function normalizeProbeTarget(item) {
  return {
    name: String(item?.name || '').trim(),
    url: String(item?.url || '').trim(),
    status_code: Number(item?.status_code) || 0,
    latency_ms: Number(item?.latency_ms) || 0,
    success: item?.success === true,
    error: String(item?.error || '').trim(),
  };
}

function normalizeProbeResult(item) {
  if (!item) {
    return null;
  }

  return {
    exit_ip: String(item.exit_ip || '').trim(),
    city: String(item.city || '').trim(),
    region: String(item.region || '').trim(),
    country: String(item.country || '').trim(),
    country_code: String(item.country_code || '').trim(),
    average_latency_ms: Number(item.average_latency_ms) || 0,
    quality: String(item.quality || '').trim(),
    quality_score: Number(item.quality_score) || 0,
    success_count: Number(item.success_count) || 0,
    failure_count: Number(item.failure_count) || 0,
    probed_at: Number(item.probed_at) || 0,
    targets: Array.isArray(item.targets)
      ? item.targets.map((target) => normalizeProbeTarget(target))
      : [],
  };
}

function normalizeProxyItem(item) {
  return {
    ui_id: item.ui_id || item.id || createLocalId(),
    id: String(item.id || '').trim(),
    name: String(item.name || '').trim(),
    proxy_url: String(item.proxy_url || '').trim(),
    usage_count: Number(item.usage_count) || 0,
    probe: normalizeProbeResult(item.probe),
  };
}

function stripProxyForSave(item) {
  return {
    id: String(item.id || '').trim(),
    name: String(item.name || '').trim(),
    proxy_url: String(item.proxy_url || '').trim(),
  };
}

function isProbeReady(item) {
  return (
    String(item.name || '').trim() !== '' &&
    String(item.proxy_url || '').trim() !== ''
  );
}

function formatLatency(value) {
  const latency = Number(value) || 0;
  if (latency <= 0) {
    return '超时';
  }
  return `${latency} ms`;
}

function formatProbeTime(value) {
  const ts = Number(value) || 0;
  if (!ts) {
    return '';
  }
  return new Date(ts * 1000).toLocaleString();
}

const ProxyPoolSetting = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [probing, setProbing] = useState(false);
  const [proxies, setProxies] = useState([]);
  const originSnapshotRef = useRef('[]');

  const probeProxyPools = async (baseItems = proxies, options = {}) => {
    const { silent = false } = options;
    const probeReadyIndexes = [];
    const payload = [];

    baseItems.forEach((item, index) => {
      if (!isProbeReady(item)) {
        return;
      }
      probeReadyIndexes.push(index);
      payload.push(stripProxyForSave(item));
    });

    if (payload.length === 0) {
      setProxies((prev) => prev.map((item) => ({ ...item, probe: null })));
      return;
    }

    setProbing(true);
    try {
      const res = await API.post('/api/channel/proxy_pools/probe', {
        proxies: payload,
      });
      const { success, message, data } = res.data;
      if (!success) {
        if (!silent) {
          showError(message);
        }
        return;
      }

      const items = Array.isArray(data?.items) ? data.items : [];
      setProxies((prev) => {
        const merged = prev.map((item) => ({ ...normalizeProxyItem(item) }));
        probeReadyIndexes.forEach((baseIndex, responseIndex) => {
          if (!merged[baseIndex]) {
            return;
          }
          merged[baseIndex] = {
            ...merged[baseIndex],
            probe: normalizeProbeResult(items[responseIndex]?.probe),
          };
        });
        return merged;
      });
      if (!silent) {
        showSuccess(t('探测完成'));
      }
    } catch (error) {
      if (!silent) {
        showError(t('探测失败'));
      }
    } finally {
      setProbing(false);
    }
  };

  const loadProxyPools = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/channel/proxy_pools');
      const { success, message, data } = res.data;
      if (!success) {
        showError(message);
        return;
      }

      const items = Array.isArray(data?.items) ? data.items : [];
      const normalized = items.map((item) => normalizeProxyItem(item));
      setProxies(normalized);
      originSnapshotRef.current = JSON.stringify(
        normalized.map(({ id, name, proxy_url }) => ({ id, name, proxy_url })),
      );
      await probeProxyPools(normalized, { silent: true });
    } catch (error) {
      showError(t('加载代理池失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProxyPools();
  }, []);

  const addProxy = () => {
    setProxies((prev) => [
      ...prev,
      normalizeProxyItem({
        ui_id: createLocalId(),
        id: '',
        name: '',
        proxy_url: '',
      }),
    ]);
  };

  const removeProxy = (uiId) => {
    setProxies((prev) => prev.filter((item) => item.ui_id !== uiId));
  };

  const updateProxy = (uiId, key, value) => {
    setProxies((prev) =>
      prev.map((item) =>
        item.ui_id === uiId
          ? {
              ...item,
              [key]: value,
              ...(key === 'proxy_url' ? { probe: null } : {}),
            }
          : item,
      ),
    );
  };

  const saveProxyPools = async () => {
    const payload = proxies
      .map((item) => stripProxyForSave(item))
      .filter((item) => item.name !== '' || item.proxy_url !== '');

    const snapshot = JSON.stringify(payload);
    if (snapshot === originSnapshotRef.current) {
      showWarning(t('你似乎并没有修改什么'));
      return;
    }

    setSaving(true);
    try {
      const res = await API.put('/api/option/proxy_pools', {
        proxies: payload,
      });
      const { success, message, data } = res.data;
      if (!success) {
        showError(message);
        return;
      }

      const items = Array.isArray(data?.items) ? data.items : [];
      const normalized = items.map((item) => normalizeProxyItem(item));
      setProxies(normalized);
      originSnapshotRef.current = JSON.stringify(
        normalized.map(({ id, name, proxy_url }) => ({ id, name, proxy_url })),
      );
      showSuccess(t('保存成功'));
      await probeProxyPools(normalized, { silent: true });
    } catch (error) {
      showError(t('保存失败'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Spin spinning={loading || saving || probing}>
      <div className='space-y-4'>
        <Banner
          type='info'
          description={t(
            '这里维护全局代理池，页面打开时会自动探测一次，也可以手动重新探测，不会对容器中的其他请求自动生效。',
          )}
        />

        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div>
            <Text className='text-lg font-medium'>{t('代理池列表')}</Text>
            <div className='text-xs text-gray-500'>
              {t('使用稳定的 ID 作为渠道引用，重命名不会影响已绑定的渠道。')}
            </div>
          </div>
          <div className='flex flex-wrap gap-2'>
            <Button icon={<IconRefresh />} onClick={loadProxyPools}>
              {t('刷新')}
            </Button>
            <Button
              icon={<IconRefresh />}
              onClick={() => probeProxyPools(proxies)}
              disabled={proxies.length === 0}
            >
              {t('重新探测')}
            </Button>
            <Button icon={<IconPlus />} onClick={addProxy} theme='outline'>
              {t('新增代理')}
            </Button>
            <Button icon={<IconSave />} type='primary' onClick={saveProxyPools}>
              {t('保存代理池')}
            </Button>
          </div>
        </div>

        {proxies.length === 0 ? (
          <Banner
            type='tip'
            description={t('当前还没有代理池，点击“新增代理”开始添加。')}
          />
        ) : (
          <div className='space-y-3'>
            {proxies.map((item) => (
              <div
                key={item.ui_id}
                className='rounded-xl border border-gray-200 bg-white p-4 shadow-sm'
              >
                <div className='flex items-start justify-between gap-4'>
                  <div className='min-w-0'>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Text className='text-base font-medium'>
                        {item.name || t('未命名代理')}
                      </Text>
                      <Tag color='blue'>
                        {t('引用 {{count}} 个渠道', {
                          count: item.usage_count || 0,
                        })}
                      </Tag>
                    </div>
                    <Text className='mt-1 block text-xs text-gray-500 font-mono break-all'>
                      {item.id || t('保存后自动生成 ID')}
                    </Text>
                  </div>
                  <Button
                    type='danger'
                    theme='borderless'
                    icon={<IconDelete />}
                    onClick={() => removeProxy(item.ui_id)}
                  />
                </div>

                <Row gutter={12} className='mt-3'>
                  <Col xs={24} md={8}>
                    <Text className='mb-1 block text-xs text-gray-500'>
                      {t('名称')}
                    </Text>
                    <Input
                      value={item.name}
                      onChange={(value) => updateProxy(item.ui_id, 'name', value)}
                      placeholder={t('例如：香港 01')}
                      showClear
                    />
                  </Col>
                  <Col xs={24} md={16}>
                    <Text className='mb-1 block text-xs text-gray-500'>
                      {t('代理地址')}
                    </Text>
                    <Input
                      value={item.proxy_url}
                      onChange={(value) =>
                        updateProxy(item.ui_id, 'proxy_url', value)
                      }
                      placeholder='socks5://user:pass@host:port'
                      showClear
                    />
                  </Col>
                </Row>

                <div className='mt-4 rounded-xl bg-gray-50 p-3'>
                  <div className='flex flex-wrap items-center justify-between gap-2'>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Text className='text-xs text-gray-500'>
                        {t('探测结果')}
                      </Text>
                      <Tag
                        color={
                          QUALITY_COLORS[item.probe?.quality] || QUALITY_COLORS.未知
                        }
                      >
                        {item.probe?.quality || t('未探测')}
                      </Tag>
                      {typeof item.probe?.quality_score === 'number' && (
                        <Tag color='blue'>
                          {t('评分 {{score}}', {
                            score: item.probe.quality_score,
                          })}
                        </Tag>
                      )}
                    </div>
                    {item.probe?.probed_at ? (
                      <Text className='text-xs text-gray-500'>
                        {t('最近探测')}: {formatProbeTime(item.probe.probed_at)}
                      </Text>
                    ) : null}
                  </div>

                  <div className='mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4'>
                    <div>
                      <Text className='text-xs text-gray-500'>
                        {t('出口 IP')}
                      </Text>
                      <div className='font-mono text-sm break-all'>
                        {item.probe?.exit_ip || t('未知')}
                      </div>
                    </div>
                    <div>
                      <Text className='text-xs text-gray-500'>
                        {t('所在城市')}
                      </Text>
                      <div className='text-sm break-all'>
                        {item.probe?.city || t('未知')}
                        {item.probe?.country
                          ? ` · ${item.probe.country}`
                          : ''}
                      </div>
                    </div>
                    <div>
                      <Text className='text-xs text-gray-500'>
                        {t('平均延迟')}
                      </Text>
                      <div className='text-sm'>
                        {formatLatency(item.probe?.average_latency_ms)}
                      </div>
                    </div>
                    <div>
                      <Text className='text-xs text-gray-500'>
                        {t('成功 / 失败')}
                      </Text>
                      <div className='text-sm'>
                        {item.probe
                          ? `${item.probe.success_count || 0} / ${item.probe.failure_count || 0}`
                          : t('未探测')}
                      </div>
                    </div>
                  </div>

                  {Array.isArray(item.probe?.targets) &&
                  item.probe.targets.length > 0 ? (
                    <div className='mt-3 flex flex-wrap gap-2'>
                      {item.probe.targets.map((target) => {
                        const node = (
                          <Tag
                            key={`${item.ui_id}-${target.name}`}
                            color={target.success ? 'green' : 'red'}
                          >
                            {target.name} {formatLatency(target.latency_ms)}
                          </Tag>
                        );

                        if (!target.error) {
                          return node;
                        }

                        return (
                          <Tooltip
                            key={`${item.ui_id}-${target.name}`}
                            content={target.error}
                          >
                            {node}
                          </Tooltip>
                        );
                      })}
                    </div>
                  ) : (
                    <div className='mt-3 text-xs text-gray-400'>
                      {t('页面加载后会自动探测，探测结果会显示在这里。')}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Spin>
  );
};

export default ProxyPoolSetting;
