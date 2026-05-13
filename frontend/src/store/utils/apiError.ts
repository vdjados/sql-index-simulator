import axios from 'axios'

export function apiErrMessage(e: unknown): string {
  if (axios.isAxiosError(e)) {
    const d = e.response?.data
    if (d && typeof d === 'object' && 'description' in d) {
      return String((d as { description?: string }).description ?? 'Ошибка запроса')
    }
  }
  return 'Ошибка запроса'
}
