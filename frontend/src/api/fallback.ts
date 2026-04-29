import { MOCK_INDEXED_TABLES } from '../mock/indexedTables'
import type { ApiService } from './client'

export function fallbackServices(filter?: string): ApiService[] {
  const q = filter?.trim().toLowerCase() ?? ''
  return MOCK_INDEXED_TABLES.filter((x) => {
    if (!q) return true
    return x.name.toLowerCase().includes(q) || x.tableSize.toLowerCase().includes(q)
  }).map((x) => ({
    id: x.id,
    name: x.name,
    table_size: x.tableSize,
    speed: x.speed,
    // Для CLIP обязательно английское описание — поэтому в fallback берём то,
    // что есть (если бэкенда нет, CLIP-поиск можно всё равно попробовать на этом).
    description: x.description,
    short_description_en: x.description,
    image_key: '',
    video_key: '',
    image_url: x.imageUrl,
    video_url: x.gifUrl,
  }))
}

export function fallbackServiceById(id: string): ApiService | undefined {
  const x = MOCK_INDEXED_TABLES.find((t) => t.id === id)
  if (!x) return undefined
  return {
    id: x.id,
    name: x.name,
    table_size: x.tableSize,
    speed: x.speed,
    description: x.description,
    short_description_en: x.description,
    image_key: '',
    video_key: '',
    image_url: x.imageUrl,
    video_url: x.gifUrl,
  }
}

