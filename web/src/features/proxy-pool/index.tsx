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

For commercial licensing, please support contact@quantumnous.com
*/
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Globe, Plus, Save, Search, Trash2, WifiOff } from 'lucide-react'
import { useCallback, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'

import { getProxyPools, probeProxyPools, saveProxyPools } from './api'
import type { ProxyPoolItem } from './types'

export function ProxyPoolPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['proxy-pools'],
    queryFn: getProxyPools,
  })

  const items = useMemo(() => data?.data?.items ?? [], [data])

  const [editingItems, setEditingItems] = useState<ProxyPoolItem[]>([])
  const [isDirty, setIsDirty] = useState(false)
  const [probeResults, setProbeResults] = useState<
    Record<string, { success: boolean; latency_ms: number; error?: string }>
  >({})

  // Initialize editing items from data
  const initEditingItems = useCallback(() => {
    setEditingItems(
      items.map((item) => ({
        id: item.id,
        name: item.name,
        proxy_url: item.proxy_url,
      }))
    )
    setProbeResults({})
    setIsDirty(false)
  }, [items])

  // Sync when data loads
  const [initialized, setInitialized] = useState(false)
  if (!initialized && items.length > 0) {
    initEditingItems()
    setInitialized(true)
  }

  const saveMutation = useMutation({
    mutationFn: (proxies: ProxyPoolItem[]) => saveProxyPools(proxies),
    onSuccess: (res) => {
      if (res.success) {
        toast.success(t('Proxy pool saved successfully'))
        queryClient.invalidateQueries({ queryKey: ['proxy-pools'] })
        setIsDirty(false)
        setProbeResults({})
      }
    },
  })

  const probeMutation = useMutation({
    mutationFn: (proxies: ProxyPoolItem[]) => probeProxyPools(proxies),
    onSuccess: (res) => {
      if (res.success && res.data?.items) {
        const results: Record<string, { success: boolean; latency_ms: number; error?: string }> = {}
        for (const item of res.data.items) {
          if (item.probe) {
            results[item.id] = item.probe
          }
        }
        setProbeResults(results)
        const successCount = res.data.items.filter((i) => i.probe?.success).length
        toast.success(
          t('Probe complete: {{success}}/{{total}} succeeded', {
            success: successCount,
            total: res.data.items.length,
          })
        )
      }
    },
  })

  const handleAdd = useCallback(() => {
    setEditingItems((prev) => [
      ...prev,
      { id: '', name: '', proxy_url: '' },
    ])
    setIsDirty(true)
  }, [])

  const handleRemove = useCallback((index: number) => {
    setEditingItems((prev) => prev.filter((_, i) => i !== index))
    setIsDirty(true)
  }, [])

  const handleChange = useCallback(
    (index: number, field: 'name' | 'proxy_url', value: string) => {
      setEditingItems((prev) => {
        const next = [...prev]
        next[index] = { ...next[index], [field]: value }
        return next
      })
      setIsDirty(true)
    },
    []
  )

  const handleSave = useCallback(() => {
    const validItems = editingItems.filter(
      (item) => item.name.trim() || item.proxy_url.trim()
    )
    saveMutation.mutate(validItems)
  }, [editingItems, saveMutation])

  const handleProbe = useCallback(() => {
    const validItems = editingItems.filter(
      (item) => item.name.trim() && item.proxy_url.trim()
    )
    if (validItems.length === 0) {
      toast.error(t('No valid proxy items to probe'))
      return
    }
    probeMutation.mutate(validItems)
  }, [editingItems, probeMutation])

  return (
    <div className='mx-auto max-w-5xl space-y-6 p-4 sm:p-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>
            {t('Proxy Pool')}
          </h1>
          <p className='text-muted-foreground text-sm'>
            {t('Configure proxy servers for channel routing')}
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button
            variant='outline'
            onClick={handleProbe}
            disabled={probeMutation.isPending || editingItems.length === 0}
          >
            <Search className='mr-2 size-4' />
            {probeMutation.isPending ? t('Probing...') : t('Probe All')}
          </Button>
          <Button
            onClick={handleSave}
            disabled={saveMutation.isPending || !isDirty}
          >
            <Save className='mr-2 size-4' />
<Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-3'>
          <div>
            <CardTitle className='text-base'>{t('Proxy Servers')}</CardTitle>
            <CardDescription>
              {t('Add, edit, or remove proxy servers. Proxies are used by channels with Proxy Mode set to "Pool".')}
            </CardDescription>
          </div>
          <Button variant='outline' size='sm' onClick={handleAdd}>
            <Plus className='mr-2 size-4' />
            {t('Add Proxy')}
          </Button>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className='text-muted-foreground py-8 text-center text-sm'>
              {t('Loading...')}
            </div>
          ) : editingItems.length === 0 ? (
            <div className='text-muted-foreground py-8 text-center text-sm'>
              <Globe className='mx-auto mb-2 size-8 opacity-50' />
              {t('No proxy servers configured. Click "Add Proxy" to get started.')}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className='w-[40%]'>{t('Name')}</TableHead>
                  <TableHead className='w-[40%]'>{t('Proxy URL')}</TableHead>
                  <TableHead className='w-[10%] text-center'>
                    {t('Usage')}
                  </TableHead>
                  <TableHead className='w-[10%] text-center'>
                    {t('Latency')}
                  </TableHead>
                  <TableHead className='w-[1%]' />
                </TableRow>
              </TableHeader>
              <TableBody>
                {editingItems.map((item, index) => {
                  const originalItem = items.find(
                    (i) => i.id === item.id
                  )
                  const usageCount = originalItem?.usage_count ?? 0

                  return (
                    <TableRow key={index}>
                      <TableCell>
                        <Input
                          value={item.name}
                          onChange={(e) =>
                            handleChange(index, 'name', e.target.value)
                          }
                          placeholder={t('Proxy name')}
                          className='h-8'
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          value={item.proxy_url}
                          onChange={(e) =>
                            handleChange(index, 'proxy_url', e.target.value)
                          }
                          placeholder='http://... or socks5://...'
                          className='h-8 font-mono text-xs'
                        />
                      </TableCell>
                      <TableCell className='text-center text-sm tabular-nums'>
                        {usageCount}
                      </TableCell>
                      <TableCell className='text-center'>
                        {probeResults[item.id] ? (
                          probeResults[item.id].success ? (
                            <span className='text-success text-xs font-medium tabular-nums'>
                              {probeResults[item.id].latency_ms}ms
                            </span>
                          ) : (
                            <span className='text-destructive flex items-center justify-center gap-1 text-xs'>
                              <WifiOff className='size-3' />
                              {probeResults[item.id].error || t('Failed')}
                            </span>
                          )
                        ) : originalItem?.probe ? (
                          originalItem.probe.success ? (
                            <span className='text-success text-xs font-medium tabular-nums'>
                              {originalItem.probe.latency_ms}ms
                            </span>
                          ) : (
                            <span className='text-destructive flex items-center justify-center gap-1 text-xs'>
                              <WifiOff className='size-3' />
                              {originalItem.probe.error || t('Failed')}
                            </span>
                          )
                        ) : (
                          <span className='text-muted-foreground text-xs'>
                            -
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='size-8 text-destructive'
                          onClick={() => handleRemove(index)}
                        >
                          <Trash2 className='size-4' />
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}