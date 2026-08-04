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
import { api } from '@/lib/api'

import type {
  ProxyPoolItem,
  ProxyPoolListResponse,
} from './types'

/**
 * Fetch all proxy pools
 */
export async function getProxyPools(): Promise<{
  success: boolean
  message?: string
  data: ProxyPoolListResponse
}> {
  const res = await api.get('/api/proxy_pool/')
  return res.data
}

/**
 * Save proxy pools
 */
export async function saveProxyPools(
  proxies: ProxyPoolItem[]
): Promise<{
  success: boolean
  message?: string
  data: ProxyPoolListResponse
}> {
  const res = await api.post('/api/proxy_pool/', { proxies })
  return res.data
}

/**
 * Probe proxy pools
 */
export async function probeProxyPools(
  proxies: ProxyPoolItem[]
): Promise<{
  success: boolean
  message?: string
  data: ProxyPoolListResponse
}> {
  const res = await api.post('/api/proxy_pool/probe', { proxies })
  return res.data
}