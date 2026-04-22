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
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconDelete,
  IconPlus,
  IconRefresh,
  IconSave,
} from '@douyinfe/semi-icons';

const { Text } = Typography;

function createLocalId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `proxy-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function normalizeProxyItem(item) {
  return {
    ui_id: item.ui_id || item.id || createLocalId(),
    id: String(item.id || '').trim(),
    name: String(item.name || '').trim(),
    proxy_url: String(item.proxy_url || '').trim(),
    usage_count: Number(item.usage_count) || 0,
  };
}

const ProxyPoolSetting = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [proxies, setProxies] = useState([]);
  const originSnapshotRef = useRef('[]');

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
      normalizeProxyItem({ ui_id: createLocalId(), id: '', name: '', proxy_url: '' }),
    ]);
  };

  const removeProxy = (uiId) => {
    setProxies((prev) => prev.filter((item) => item.ui_id !== uiId));
  };

  const updateProxy = (uiId, key, value) => {
    setProxies((prev) =>
      prev.map((item) =>
        item.ui_id === uiId ? { ...item, [key]: value } : item,
      ),
    );
  };

  const saveProxyPools = async () => {
    const payload = proxies
      .map((item) => ({
        id: String(item.id || '').trim(),
        name: String(item.name || '').trim(),
        proxy_url: String(item.proxy_url || '').trim(),
      }))
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
    } catch (error) {
      showError(t('保存失败'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Spin spinning={loading || saving}>
      <div className='space-y-4'>
        <Banner
          type='info'
          description={t(
            '这里维护全局代理池，渠道在编辑时手动选择后才会使用，不会对容器中的其他请求自动生效。',
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
              </div>
            ))}
          </div>
        )}
      </div>
    </Spin>
  );
};

export default ProxyPoolSetting;
