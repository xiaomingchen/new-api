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
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useMemo, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import * as z from 'zod'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'

import { SettingsForm } from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'
import { safeNumberFieldProps } from '../utils/numeric-field'

const transportSchema = z.object({
  transport_setting: z.object({
    max_idle_conns: z.coerce.number().min(0),
    max_idle_conns_per_host: z.coerce.number().min(0),
    idle_conn_timeout: z.coerce.number().min(0),
  }),
})

type TransportFormInput = z.input<typeof transportSchema>
type TransportFormValues = z.output<typeof transportSchema>

type FlatTransportDefaults = {
  'transport_setting.max_idle_conns': number
  'transport_setting.max_idle_conns_per_host': number
  'transport_setting.idle_conn_timeout': number
}

type TransportSettingsSectionProps = {
  defaultValues: FlatTransportDefaults
}

const buildFormDefaults = (
  defaults: TransportSettingsSectionProps['defaultValues']
): TransportFormInput => ({
  transport_setting: {
    max_idle_conns: defaults['transport_setting.max_idle_conns'] ?? 0,
    max_idle_conns_per_host: defaults['transport_setting.max_idle_conns_per_host'] ?? 0,
    idle_conn_timeout: defaults['transport_setting.idle_conn_timeout'] ?? 0,
  },
})

const normalizeDefaults = (
  defaults: TransportSettingsSectionProps['defaultValues']
): FlatTransportDefaults => ({
  'transport_setting.max_idle_conns': defaults['transport_setting.max_idle_conns'] ?? 0,
  'transport_setting.max_idle_conns_per_host': defaults['transport_setting.max_idle_conns_per_host'] ?? 0,
  'transport_setting.idle_conn_timeout': defaults['transport_setting.idle_conn_timeout'] ?? 0,
})

function submitTransportSettings(
  values: TransportFormValues,
  updateOption: ReturnType<typeof import('../hooks/use-update-option').useUpdateOption>,
  t: ReturnType<typeof import('react-i18next').useTranslation>['t']
) {
  const items: Array<{ key: string; value: string }> = []
  if (values.transport_setting?.max_idle_conns !== undefined) {
    items.push({ key: 'transport_setting.max_idle_conns', value: String(values.transport_setting.max_idle_conns) })
  }
  if (values.transport_setting?.max_idle_conns_per_host !== undefined) {
    items.push({ key: 'transport_setting.max_idle_conns_per_host', value: String(values.transport_setting.max_idle_conns_per_host) })
  }
  if (values.transport_setting?.idle_conn_timeout !== undefined) {
    items.push({ key: 'transport_setting.idle_conn_timeout', value: String(values.transport_setting.idle_conn_timeout) })
  }
  return Promise.allSettled(
    items.map((item) => updateOption.mutateAsync({ key: item.key, value: item.value }))
  ).then((results) => {
    const failed = results.filter((r): r is PromiseRejectedResult => r.status === 'rejected')
    if (failed.length > 0) {
      toast.error(t('Failed to save transport settings'))
      return
    }
    toast.success(t('Transport settings saved'))
  })
}

export function TransportSettingsSection({
  defaultValues,
}: TransportSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const normalizedDefaults = useMemo(() => normalizeDefaults(defaultValues), [defaultValues])

  const form = useForm<TransportFormInput, unknown, TransportFormValues>({
    resolver: zodResolver(transportSchema),
    defaultValues: useMemo(() => buildFormDefaults(defaultValues), []),
  })

  const { reset, handleSubmit, formState } = form
  const resetRef = useResetForm(reset, normalizedDefaults, buildFormDefaults)
  const prevActionsRef = useRef<SettingsPageFormActions | null>(null)

  useEffect(() => {
    if (resetRef.current) resetRef.current(normalizedDefaults)
  }, [normalizedDefaults, resetRef])

  const onSubmit = handleSubmit(async (values) => {
    await submitTransportSettings(values, updateOption, t)
  })

  return (
    <SettingsSection
      title={t('HTTP Transport Pool')}
      description={t('Configure idle connection pool sizing for upstream relay HTTP clients. Set to 0 to use environment defaults.')}
      formActions={prevActionsRef}
      onReset={() => resetRef.current?.(normalizedDefaults)}
      isDirty={formState.isDirty}
    >
      <Form {...form}>
        <SettingsForm onSubmit={onSubmit} id='transport-settings-form'>
          <div className='space-y-4'>
            <FormField
              control={form.control}
              name='transport_setting.max_idle_conns'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Max Idle Connections')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={0} step={1} {...safeNumberFieldProps(field)} />
                  </FormControl>
                  <FormDescription>{t('Maximum idle connections across all hosts. Default: 500.')}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='transport_setting.max_idle_conns_per_host'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Max Idle Connections Per Host')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={0} step={1} {...safeNumberFieldProps(field)} />
                  </FormControl>
                  <FormDescription>{t('Maximum idle connections per upstream host. Default: 100.')}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='transport_setting.idle_conn_timeout'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Idle Connection Timeout (seconds)')}</FormLabel>
                  <FormControl>
                    <Input type='number' min={0} step={1} {...safeNumberFieldProps(field)} />
                  </FormControl>
                  <FormDescription>{t('How long an idle connection is kept alive. Default: 90 seconds.')}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
