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
import { Activity, Link2, RadioTower, Wallet } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { IconBadge, type IconBadgeTone } from '@/components/ui/icon-badge'
import { Skeleton } from '@/components/ui/skeleton'
import { getChannels } from '@/features/channels/api'
import {
  formatRelativeTime,
  sortChannelsByActivity,
} from '@/features/channels/lib'
import { formatQuotaWithCurrency } from '@/lib/currency'
import { cn } from '@/lib/utils'

const CHANNELS_PAGE_SIZE = 100
const TOP_ACTIVE_LIMIT = 6

export function ChannelStatsPanel() {
  const { t, i18n } = useTranslation()
  const channelsQuery = useQuery({
    queryKey: ['dashboard-channels-stats'],
    queryFn: () => getChannels({ p: 0, page_size: CHANNELS_PAGE_SIZE }),
    staleTime: 30 * 1000,
    retry: false,
  })

  const channels = useMemo(
    () => channelsQuery.data?.data?.items ?? [],
    [channelsQuery.data]
  )

  const summary = useMemo(() => {
    let activeConnections = 0
    let inUseCount = 0
    let todayCost = 0
    let todayTokens = 0
    for (const channel of channels) {
      const connections = Number(channel.current_connections ?? 0)
      activeConnections += connections
      if (connections > 0) inUseCount += 1
      todayCost += Number(channel.used_quota_today ?? 0)
      todayTokens += Number(channel.used_tokens_today ?? 0)
    }
    return { activeConnections, inUseCount, todayCost, todayTokens }
  }, [channels])

  const topActive = useMemo(
    () => sortChannelsByActivity(channels).slice(0, TOP_ACTIVE_LIMIT),
    [channels]
  )

  const loading = channelsQuery.isLoading
  const locale = i18n.resolvedLanguage || i18n.language

  return (
    <section className='bg-card h-full overflow-hidden rounded-2xl border shadow-xs'>
      <div className='flex items-center gap-2 border-b px-4 py-3 sm:px-5'>
        <IconBadge tone='info' size='sm'>
          <RadioTower />
        </IconBadge>
        <h3 className='text-sm font-semibold'>{t('Channel runtime')}</h3>
        <span className='text-muted-foreground ml-auto text-xs'>
          {t('Live channel usage snapshot')}
        </span>
      </div>
      <div className='space-y-3 p-4 sm:p-5'>
        <div className='grid grid-cols-2 gap-2 sm:grid-cols-4'>
          <MetricCell icon={Activity} label={t('In Use')} value={summary.inUseCount.toLocaleString()} loading={loading} tone='success' />
          <MetricCell icon={RadioTower} label={t('Active Connections')} value={summary.activeConnections.toLocaleString()} loading={loading} tone='info' />
          <MetricCell icon={Wallet} label={t('Today Spend')} value={formatQuotaWithCurrency(summary.todayCost)} loading={loading} tone='warning' />
          <MetricCell icon={Link2} label={t('Today Tokens')} value={summary.todayTokens.toLocaleString()} loading={loading} tone='info' />
        </div>

        {loading ? (
          <div className='space-y-1'>
            {['a', 'b', 'c'].map((key) => (
              <Skeleton key={key} className='h-5 w-full rounded' />
            ))}
          </div>
        ) : (
          topActive.length > 0 && (
            <div>
              <span className='text-muted-foreground mb-1 block text-[11px] font-medium'>
                {t('Active channels')}
              </span>
              <div className='grid grid-cols-1 gap-x-4 sm:grid-cols-2'>
                {topActive.map((channel) => {
                  const connections = Number(channel.current_connections ?? 0)
                  return (
                    <div
                      key={channel.id}
                      className='flex items-center justify-between gap-2 rounded px-1.5 py-1'
                    >
                      <span className='min-w-0 flex-1 truncate text-[11px] font-medium'>
                        {channel.name}
                      </span>
                      <span className='inline-flex shrink-0 items-center gap-1.5'>
                        <span
                          className={cn(
                            'size-1.5 rounded-full',
                            connections > 0
                              ? 'bg-success'
                              : 'bg-muted-foreground/40'
                          )}
                          aria-hidden='true'
                        />
                        <span
                          className={cn(
                            'font-mono text-[11px] font-semibold tabular-nums',
                            connections > 0
                              ? 'text-success'
                              : 'text-muted-foreground'
                          )}
                        >
                          {connections.toLocaleString()}
                        </span>
                        <span className='text-muted-foreground font-mono text-[10px] tabular-nums'>
                          {formatRelativeTime(channel.last_used_at, locale)}
                        </span>
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          )
        )}
      </div>
    </section>
  )
}
function MetricCell(props: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string
  loading: boolean
  tone: IconBadgeTone
}) {
  const Icon = props.icon
  return (
    <div className='bg-muted/40 rounded-xl px-3 py-2.5'>
      <div className='text-muted-foreground flex items-center gap-1.5 text-[11px] font-medium'>
        <IconBadge tone={props.tone} size='xs'>
          <Icon />
        </IconBadge>
        <span className='truncate'>{props.label}</span>
      </div>
      {props.loading ? (
        <Skeleton className='mt-1.5 h-5 w-16' />
      ) : (
        <div className='mt-1.5 font-mono text-sm font-semibold tabular-nums'>
          {props.value}
        </div>
      )}
    </div>
  )
}

