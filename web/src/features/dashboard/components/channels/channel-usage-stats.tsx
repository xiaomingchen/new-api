/*
Copyright (C) 2023-2026 QuantumNous

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
import { useQuery } from '@tanstack/react-query'
import { Activity, AlertTriangle, CheckCircle, DollarSign, FileText, Hash } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { getChannelStats } from '@/features/dashboard/api'
import { formatQuotaWithCurrency } from '@/lib/currency'
import { formatTokens } from '@/lib/format'

import type { ChannelStatsItem } from '../../types'

const COLUMNS = [
  { key: 'channel_name', icon: Hash, labelKey: 'Channel Name' },
  { key: 'model_name', icon: FileText, labelKey: 'Model' },
  { key: 'request_count', icon: Activity, labelKey: 'Requests' },
  { key: 'used_tokens', icon: FileText, labelKey: 'Tokens' },
  { key: 'success_rate', icon: CheckCircle, labelKey: 'Success Rate' },
  { key: 'today_amount', icon: DollarSign, labelKey: 'Today Amount' },
  { key: 'total_amount', icon: DollarSign, labelKey: 'Total Amount' },
] as const

function formatValue(item: ChannelStatsItem, key: string): string {
  switch (key) {
    case 'channel_name':
      return item.channel_name
    case 'model_name':
      return item.model_name || '-'
    case 'request_count':
      return formatTokens(item.request_count)
    case 'used_tokens':
      return formatTokens(item.used_tokens)
    case 'success_rate':
      return `${item.success_rate.toFixed(1)}%`
    case 'today_amount':
      return formatQuotaWithCurrency(item.today_amount)
    case 'total_amount':
      return formatQuotaWithCurrency(item.total_amount)
    default:
      return ''
  }
}

function getCellClass(key: string): string {
  switch (key) {
    case 'request_count':
    case 'used_tokens':
    case 'today_amount':
    case 'total_amount':
      return 'font-mono tabular-nums text-right'
    case 'success_rate':
      return 'font-mono tabular-nums text-center'
    default:
      return ''
  }
}

function getSuccessRateColor(rate: number): string {
  if (rate >= 99) return 'text-success'
  if (rate >= 95) return 'text-warning'
  return 'text-destructive'
}

export function ChannelUsageStats() {
  const { t } = useTranslation()
  const { data, isLoading, isError } = useQuery({
    queryKey: ['channel-usage-stats'],
    queryFn: () => getChannelStats(),
    staleTime: 30 * 1000,
    retry: false,
    refetchInterval: 60 * 1000,
  })

  const stats = data?.data ?? []

  if (isLoading) {
    return (
      <div className='overflow-hidden rounded-lg border'>
        <div className='border-b px-4 py-3 sm:px-5'>
          <Skeleton className='h-5 w-40' />
        </div>
        <div className='space-y-2 p-4 sm:p-5'>
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className='h-10 w-full rounded' />
          ))}
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className='flex items-center justify-center gap-2 rounded-lg border px-4 py-12 text-sm text-red-600'>
        <AlertTriangle className='size-4' />
        <span>{t('Failed to load channel usage statistics')}</span>
      </div>
    )
  }

  if (stats.length === 0) {
    return (
      <div className='flex items-center justify-center rounded-lg border px-4 py-12 text-sm text-muted-foreground'>
        {t('No channel usage data available for today')}
      </div>
    )
  }

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='overflow-x-auto'>
        <table className='w-full text-left text-sm'>
          <thead>
            <tr className='bg-muted/50 border-b'>
              {COLUMNS.map((col) => (
                <th
                  key={col.key}
                  className='text-muted-foreground px-4 py-3 font-medium sm:px-5'
                >
                  <div className='flex items-center gap-1.5'>
                    <col.icon className='size-3.5 shrink-0' />
                    <span className='truncate'>{t(col.labelKey)}</span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className='divide-y divide-border/60'>
            {stats.map((item) => (
              <tr
                key={item.channel_id}
                className='hover:bg-muted/30 transition-colors'
              >
                {COLUMNS.map((col) => (
                  <td
                    key={col.key}
                    className={`px-4 py-3 sm:px-5 ${getCellClass(col.key)}`}
                  >
                    {col.key === 'success_rate' ? (
                      <span className={getSuccessRateColor(item.success_rate)}>
                        {formatValue(item, col.key)}
                      </span>
                    ) : col.key === 'channel_name' ? (
                      <span className='font-medium'>{item.channel_name}</span>
                    ) : (
                      formatValue(item, col.key)
                    )}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}