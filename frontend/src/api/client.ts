const base = import.meta.env.VITE_API_BASE ?? ''

/** Базовый URL API (в dev — пустой, запросы идут на тот же origin, proxy Vite на :8082). */
export function apiUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  return `${base}${p}`
}

/** Пример подключения к вашему веб-сервису: публичный список indexed tables. */
export async function fetchIndexedTablesFromApi(): Promise<unknown> {
  const res = await fetch(apiUrl('/api/indexed-tables'), {
    method: 'GET',
    headers: { Accept: 'application/json' },
  })
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }
  return res.json()
}
