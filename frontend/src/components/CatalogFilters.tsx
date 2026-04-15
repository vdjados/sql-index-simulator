import type { CatalogFiltersState } from '../hooks/useFilteredIndexedTables'

interface CatalogFiltersProps {
  filters: CatalogFiltersState
  onChange: (next: CatalogFiltersState) => void
}

/**
 * Поиск как в прошлой лабораторной: одна строка + кнопка "Найти".
 */
export function CatalogFilters(props: CatalogFiltersProps) {
  const { filters, onChange } = props

  const patch = (partial: Partial<CatalogFiltersState>) => {
    onChange({ ...filters, ...partial })
  }

  return (
    <form className="search-form" onSubmit={(e) => e.preventDefault()}>
      <input
        type="text"
        className="search-input"
        placeholder="Поиск по названию индекса"
        value={filters.nameQuery}
        onChange={(e) => patch({ nameQuery: e.target.value })}
        autoComplete="off"
      />
      <button type="submit" className="search-btn">
        Найти
      </button>
    </form>
  )
}
