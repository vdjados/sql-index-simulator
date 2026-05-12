/** Mock-данные для карточек каталога (лаба 5: только mock, без запросов к БД для списка). */
export interface MockIndexedTable {
  id: string
  name: string
  tableSize: string
  /** Селективность запроса (0..1), ближе к предметной области. */
  selectivity: number
  /** Дата публикации для фильтра по дате (ISO YYYY-MM-DD) */
  listedDate: string
  /** Краткое англ. описание — аналог поля API `short_description_en` (для SigLIP). */
  shortDescriptionEn: string
  description: string
  /** Пустая строка — на карточке показывается placeholder */
  imageUrl: string
  /** GIF для детального просмотра; если пусто, берется дефолтный GIF */
  gifUrl: string
  /** Как в service.html — бейдж «Скорость» */
  speed: string
}

export const MOCK_INDEXED_TABLES: MockIndexedTable[] = [
  {
    id: 'btree-512mb',
    name: 'B-Tree (средняя таблица)',
    tableSize: '512 MB',
    selectivity: 0.11,
    listedDate: '2026-01-10',
    shortDescriptionEn:
      'Compact B-Tree index for fast point lookups and ordered scans.',
    description: 'Compact B-Tree index for fast point lookups and ordered scans.',
    imageUrl: '',
    gifUrl: '',
    speed: 'высокая',
  },
  {
    id: 'hash-2gb',
    name: 'Hash (большая таблица)',
    tableSize: '2 GB',
    selectivity: 0.04,
    listedDate: '2026-02-05',
    shortDescriptionEn: 'Hash index optimized for equality predicates on a single key.',
    description: 'Hash index optimized for equality predicates on a single key.',
    imageUrl: '',
    gifUrl: '',
    speed: 'очень высокая',
  },
  {
    id: 'gist-128mb',
    name: 'GiST (компактная)',
    tableSize: '128 MB',
    selectivity: 0.23,
    listedDate: '2026-03-20',
    shortDescriptionEn: 'GiST index for range queries and geometric data with good recall.',
    description: 'GiST index for range queries and geometric data with good recall.',
    imageUrl: '',
    gifUrl: '',
    speed: 'средняя',
  },
  {
    id: 'gin-1gb',
    name: 'GIN (полнотекст)',
    tableSize: '1 GB',
    selectivity: 0.08,
    listedDate: '2026-03-01',
    shortDescriptionEn: 'GIN index for full-text search and arrays, trading speed for space.',
    description: 'GIN index for full-text search and arrays, trading speed for space.',
    imageUrl: '',
    gifUrl: '',
    speed: 'высокая',
  },
]

export function getMockById(id: string): MockIndexedTable | undefined {
  return MOCK_INDEXED_TABLES.find((t) => t.id === id)
}
