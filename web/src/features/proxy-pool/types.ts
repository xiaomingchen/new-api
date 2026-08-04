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

/** Proxy pool probe result */
export interface ProxyProbeResult {
  success: boolean
  latency_ms: number
  error?: string
}

/** Proxy pool item */
export interface ProxyPoolItem {
  id: string
  name: string
  proxy_url: string
}

/** Proxy pool response item (includes usage count and optional probe) */
export interface ProxyPoolResponseItem {
  id: string
  name: string
  proxy_url: string
  usage_count: number
  probe?: ProxyProbeResult
}

/** Proxy pool list response */
export interface ProxyPoolListResponse {
  items: ProxyPoolResponseItem[]
}

/** Proxy pool save request */
export interface ProxyPoolSaveRequest {
  proxies: ProxyPoolItem[]
}

/** Proxy pool probe request */
export interface ProxyPoolProbeRequest {
  proxies: ProxyPoolItem[]
}