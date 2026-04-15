import { useMemo } from 'react'
import type { MockIndexedTable } from '../mock/indexedTables'

export interface CatalogFiltersState {
  nameQuery: string
}

export const defaultFilters: CatalogFiltersState = {
  nameQuery: '',
}

export function useFilteredIndexedTables(
  items: MockIndexedTable[],
  filters: CatalogFiltersState,
): MockIndexedTable[] {
  return useMemo(() => {
    return items.filter((row) => {
      if (filters.nameQuery.trim()) {
        const q = filters.nameQuery.trim().toLowerCase()
        if (!row.name.toLowerCase().includes(q)) {
          return false
        }
      }
      return true
    })
  }, [items, filters])
}
