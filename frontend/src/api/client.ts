const base = import.meta.env.VITE_API_BASE ?? ''

/** Базовый URL API (в dev — пустой, запросы идут на тот же origin, proxy Vite на :8082). */
export function apiUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  return `${base}${p}`
}

export interface ApiService {
  id: string
  name: string
  table_size: string
  speed: string
  description: string
  short_description_en: string
  image_key: string
  video_key: string
  image_url: string
  video_url: string
}

export interface ApiCart {
  id: number | null
  count: number
}

function safeString(v: unknown): string {
  return typeof v === 'string' ? v : ''
}

function isApiService(x: unknown): x is ApiService {
  if (!x || typeof x !== 'object') return false
  const o = x as Record<string, unknown>
  return typeof o.id === 'string' && typeof o.name === 'string'
}

async function parseJson<T>(res: Response): Promise<T> {
  const text = await res.text()
  return text ? (JSON.parse(text) as T) : (null as T)
}

async function fetchJson<T>(path: string): Promise<T> {
  const res = await fetch(apiUrl(path), {
    method: 'GET',
    headers: { Accept: 'application/json' },
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }
  return parseJson<T>(res)
}

export async function fetchServices(params: { filter?: string }): Promise<ApiService[]> {
  const url = new URL(apiUrl('/api/indexed-tables'), window.location.origin)
  if (params.filter?.trim()) url.searchParams.set('filter', params.filter.trim())
  const res = await fetch(url.toString(), {
    method: 'GET',
    headers: { Accept: 'application/json' },
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }
  const data = await parseJson<unknown>(res)
  if (!Array.isArray(data)) return []
  return data.filter(isApiService)
}

export async function fetchServiceById(id: string): Promise<ApiService> {
  const safeId = encodeURIComponent(id)
  const svc = await fetchJson<unknown>(`/api/indexed-tables/${safeId}`)
  if (!isApiService(svc)) {
    throw new Error('Unexpected service payload')
  }
  const raw = svc as unknown as Record<string, unknown>
  // Normalize empty strings.
  return {
    ...svc,
    description: safeString(svc.description),
    short_description_en: safeString(raw.short_description_en),
    image_key: safeString(svc.image_key),
    video_key: safeString(svc.video_key),
    image_url: safeString(svc.image_url),
    video_url: safeString(svc.video_url),
    table_size: safeString(svc.table_size),
    speed: safeString(svc.speed),
  }
}

/** 3-й обязательный GET: иконка корзины (без авторизации, ответ 200). */
export async function fetchPublicCart(): Promise<ApiCart> {
  const cart = await fetchJson<unknown>('/api/cart')
  if (!cart || typeof cart !== 'object') return { id: null, count: 0 }
  const o = cart as Record<string, unknown>
  const count = typeof o.count === 'number' ? o.count : 0
  const id = typeof o.id === 'number' ? o.id : null
  return { id, count }
}
