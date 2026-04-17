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

import React from 'react';
import { Button, Card, Empty, Table, Tag, Typography } from '@douyinfe/semi-ui';
import { Activity, RefreshCw } from 'lucide-react';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { renderNumber, renderQuota } from '../../helpers';

const { Text } = Typography;

const getSuccessRateColor = (successRate) => {
  if (successRate >= 99) {
    return 'green';
  }
  if (successRate >= 95) {
    return 'blue';
  }
  if (successRate >= 80) {
    return 'orange';
  }
  return 'red';
};

const ChannelStatsPanel = ({
  channelStats,
  loading,
  onRefresh,
  CARD_PROPS,
  t,
}) => {
  const columns = [
    {
      title: t('渠道'),
      dataIndex: 'channel_name',
      key: 'channel_name',
      width: 160,
      render: (value, record) => value || `${t('渠道')} #${record.channel_id}`,
    },
    {
      title: t('模型'),
      dataIndex: 'model_name',
      key: 'model_name',
      width: 180,
      render: (value) => value || t('未知'),
    },
    {
      title: t('今日调用次数'),
      dataIndex: 'request_count',
      key: 'request_count',
      width: 120,
      render: (value) => renderNumber(value || 0),
    },
    {
      title: t('当日 Token'),
      dataIndex: 'used_tokens',
      key: 'used_tokens',
      width: 120,
      render: (value) => renderNumber(value || 0),
    },
    {
      title: t('今日消费金额'),
      dataIndex: 'today_amount',
      key: 'today_amount',
      width: 140,
      render: (value) => renderQuota(value || 0),
    },
    {
      title: t('总金额'),
      dataIndex: 'total_amount',
      key: 'total_amount',
      width: 140,
      render: (value) => renderQuota(value || 0),
    },
    {
      title: t('成功调用'),
      dataIndex: 'success_count',
      key: 'success_count',
      width: 120,
      render: (value) => renderNumber(value || 0),
    },
    {
      title: t('失败调用'),
      dataIndex: 'error_count',
      key: 'error_count',
      width: 120,
      render: (value) => renderNumber(value || 0),
    },
    {
      title: t('调用成功率'),
      dataIndex: 'success_rate',
      key: 'success_rate',
      width: 140,
      render: (value) => (
        <Tag color={getSuccessRateColor(Number(value) || 0)} shape='circle'>
          {`${(Number(value) || 0).toFixed(1)}%`}
        </Tag>
      ),
    },
  ];

  return (
    <Card
      {...CARD_PROPS}
      className='!rounded-2xl'
      title={
        <div className='flex items-center gap-2'>
          <Activity size={16} />
          {t('今日渠道统计')}
        </div>
      }
      bodyStyle={{ paddingTop: 12 }}
    >
      <div className='mb-3 flex items-center justify-between gap-3'>
        <Text type='tertiary' className='flex-1'>
          {t(
            '统计范围为服务器今日 00:00 至当前时间；成功=消费日志，失败=错误日志。',
          )}
        </Text>
        <Button
          type='tertiary'
          size='small'
          icon={<RefreshCw size={14} />}
          onClick={onRefresh}
          loading={loading}
          className='!rounded-full'
        >
          {t('刷新')}
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={channelStats}
        rowKey='channel_id'
        loading={loading}
        pagination={
          channelStats.length > 10
            ? {
                pageSize: 10,
                showSizeChanger: false,
              }
            : false
        }
        scroll={{ x: 1180 }}
        empty={
          <Empty
            image={<IllustrationNoResult style={{ width: 120, height: 120 }} />}
            darkModeImage={
              <IllustrationNoResultDark style={{ width: 120, height: 120 }} />
            }
            description={t('今日暂无渠道调用数据')}
            style={{ padding: 24 }}
          />
        }
      />
    </Card>
  );
};

export default ChannelStatsPanel;
