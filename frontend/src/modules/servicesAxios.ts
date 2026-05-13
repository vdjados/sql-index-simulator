import axios from 'axios'
import type { ApiCart, ApiService } from '../api/client'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api'

/** Axios для каталога indexed-tables и публичной корзины /cart (без codegen). */
export const servicesAxios = axios.create({
  baseURL,
})

function isApiService(x: unknown): x is ApiService {
  if (!x || typeof x !== 'object') return false
  const o = x as Record<string, unknown>
  return typeof o.id === 'string' && typeof o.name === 'string'
}

export async function axiosFetchServices(params: { filter?: string }): Promise<ApiService[]> {
  const res = await servicesAxios.get<unknown[]>('/indexed-tables', {
    params: params.filter?.trim() ? { filter: params.filter.trim() } : {},
    headers: { Accept: 'application/json' },
  })
  const data = res.data
  if (!Array.isArray(data)) return []
  return data.filter(isApiService)
}

export async function axiosFetchServiceById(id: string): Promise<ApiService> {
  const safeId = encodeURIComponent(id)
  const res = await servicesAxios.get<unknown>(`/indexed-tables/${safeId}`, {
    headers: { Accept: 'application/json' },
  })
  const svc = res.data
  if (!isApiService(svc)) {
    throw new Error('Unexpected service payload')
  }
  const raw = svc as unknown as Record<string, unknown>
  return {
    ...svc,
    description: typeof svc.description === 'string' ? svc.description : '',
    short_description_en: typeof raw.short_description_en === 'string' ? raw.short_description_en : '',
    image_key: typeof raw.image_key === 'string' ? raw.image_key : '',
    video_key: typeof raw.video_key === 'string' ? raw.video_key : '',
    image_url: typeof raw.image_url === 'string' ? raw.image_url : '',
    video_url: typeof raw.video_url === 'string' ? raw.video_url : '',
    table_size: typeof svc.table_size === 'string' ? svc.table_size : '',
    speed: typeof svc.speed === 'string' ? svc.speed : '',
  }
}

export async function axiosFetchPublicCart(): Promise<ApiCart> {
  const res = await servicesAxios.get<unknown>('/cart', {
    headers: { Accept: 'application/json' },
  })
  const cart = res.data
  if (!cart || typeof cart !== 'object') return { id: null, count: 0 }
  const o = cart as Record<string, unknown>
  const count = typeof o.count === 'number' ? o.count : 0
  const id = typeof o.id === 'number' ? o.id : null
  return { id, count }
}
