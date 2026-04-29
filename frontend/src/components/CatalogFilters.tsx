import type { CatalogFiltersState } from '../hooks/useFilteredIndexedTables'

interface CatalogFiltersProps {
  filters: CatalogFiltersState
  onChange: (next: CatalogFiltersState) => void
  onImageUpload?: (file: File) => void
  onResetImageSearch?: () => void
  ready?: boolean
}

/**
 * Поиск как в прошлой лабораторной: одна строка + кнопка "Найти".
 */
export function CatalogFilters(props: CatalogFiltersProps) {
  const { filters, onChange, onImageUpload, onResetImageSearch, ready } = props

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
      <input
        type="file"
        accept="image/*"
        aria-label="Поиск по изображению"
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (file && onImageUpload) onImageUpload(file)
        }}
        style={{ maxWidth: 220 }}
        disabled={ready === false}
      />
      <button
        type="button"
        className="search-btn"
        onClick={() => onResetImageSearch?.()}
        style={{ marginLeft: 8 }}
      >
        Сбросить фото
      </button>
      <button type="submit" className="search-btn">
        Найти
      </button>
    </form>
  )
}
