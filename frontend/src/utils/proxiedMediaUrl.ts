/** В dev Vite проксирует `/media` на MinIO (см. vite.config.ts). */
export function proxiedMediaUrl(url: string): string {
  const u = url?.trim()
  if (!u) return u
  if (!import.meta.env.DEV) return u
  try {
    const parsed =
      u.startsWith('http://') || u.startsWith('https://')
        ? new URL(u)
        : new URL(u, window.location.origin)
    if (parsed.port === '9000') {
      return `/media${parsed.pathname}${parsed.search}`
    }
  } catch {
    return u
  }
  return u
}
