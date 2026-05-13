import { useRef } from 'react'
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
  const fileInputRef = useRef<HTMLInputElement>(null)

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
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="file-input-native"
        aria-label="Image search — choose file"
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (file && onImageUpload) onImageUpload(file)
          e.target.value = ''
        }}
        disabled={ready === false}
      />
      <button
        type="button"
        className="search-btn search-btn--file"
        disabled={ready === false}
        onClick={() => fileInputRef.current?.click()}
      >
        Choose image
      </button>
      <button type="button" className="search-btn search-btn--outline catalog-filters-reset" onClick={() => onResetImageSearch?.()}>
        Сбросить фото
      </button>
      <button type="submit" className="search-btn">
        Найти
      </button>
    </form>
  )
}
