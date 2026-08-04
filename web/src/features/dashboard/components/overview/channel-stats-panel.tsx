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
import { useTranslation } from 'react-i18next'

import { CardStaggerItem } from '@/components/page-transition'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'

import { getTodayChannelStats } from '../../api'
import type { ChannelStatsItem } from '../../api'

export function ChannelStatsPanel() {
  const { t } = useTranslation()

  const { data, isLoading } = useQuery({
    queryKey: ['channel-stats-today'],
    queryFn: getTodayChannelStats,
    refetchInterval: 60_000,
  })

  const stats = data?.data ?? []

  return (
    <CardStaggerItem className='lg:col-span-2'>
      <Card>
        <CardHeader className='pb-3'>
          <CardTitle className='text-base'>
            {t('Channel Stats (Today)')}
          </CardTitle>
        </CardHeader>
        <CardContent className='p-0'>
          {isLoading ? (
            <div className='text-muted-foreground py-8 text-center text-sm'>
              {t('Loading...')}
            </div>
          ) : stats.length === 0 ? (
            <div className='text-muted-foreground py-8 text-center text-sm'>
              {t('No channel activity today')}
            </div>
          ) : (
            <div className='overflow-x-auto'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('Channel')}</TableHead>
                    <TableHead>{t('Model')}</TableHead>
                    <TableHead className='text-right'>{t('Requests')}</TableHead>
                    <TableHead className='text-right'>{t('Tokens')}</TableHead>
                    <TableHead className='text-right'>{t('Today')}</TableHead>
                    <TableHead className='text-right'>{t('Total')}</TableHead>
                    <TableHead className='text-right'>{t('Success Rate')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stats.map((item: ChannelStatsItem) => (
                    <TableRow key={item.channel_id}>
                      <TableCell className='max-w-[160px] truncate font-medium'>
                        {item.channel_name}
                      </TableCell>
                      <TableCell className='text-muted-foreground text-xs'>
                        {item.model_name}
                      </TableCell>
                      <TableCell className='text-right tabular-nums'>
                        {item.request_count}
                      </TableCell>
                      <TableCell className='text-right tabular-nums'>
                        {formatQuota(item.used_tokens)}
                      </TableCell>
                      <TableCell className='text-right tabular-nums'>
                        {formatQuota(item.today_amount)}
                      </TableCell>
                      <TableCell className='text-right tabular-nums'>
                        {formatQuota(item.total_amount)}
                      </TableCell>
                      <TableCell className='text-right'>
                        <span
                          className={cn(
                            'inline-flex items-center gap-1 text-xs font-medium',
                            item.success_rate >= 99
                              ? 'text-success'
                              : item.success_rate >= 90
                                ? 'text-warning'
                                : 'text-destructive'
                          )}
                        >
                          {item.success_rate.toFixed(1)}%
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </CardStaggerItem>
  )
}