import { useEffect, useMemo, useState } from 'react'
import { CartIcon } from '../components/CartIcon'
import { CatalogFilters } from '../components/CatalogFilters'
import { IndexedTableCard } from '../components/IndexedTableCard'
import {
  defaultFilters,
  type CatalogFiltersState,
} from '../hooks/useFilteredIndexedTables'
import { axiosFetchPublicCart, axiosFetchServices } from '../modules/servicesAxios'
import { fallbackServices } from '../api/fallback'
import { useServiceImageSearch } from '../hooks/useServiceImageSearch'
import type { ApiService } from '../api/client'

/** Каталог: услуги через axios; корзина гостя через axios GET /cart. */
export function CatalogPage() {
  const [cartCount, setCartCount] = useState(0)
  const [filters, setFilters] = useState<CatalogFiltersState>(defaultFilters)
  const [items, setItems] = useState<ApiService[]>([])
  const [status, setStatus] = useState<'idle' | 'loading' | 'ok' | 'fallback'>('idle')

  const filter = filters.nameQuery.trim()
  const { items: ranked, ready, searchByImage, resetSearch, imageEmbedding } =
    useServiceImageSearch(items, { threshold: 0.4, topK: 12 })
  const visible = useMemo(() => ranked.filter((x) => x.isVisible), [ranked])

  useEffect(() => {
    let cancelled = false
    axiosFetchPublicCart()
      .then((c) => {
        if (cancelled) return
        setCartCount(c.count)
      })
      .catch(() => {
        if (cancelled) return
        setCartCount(0)
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    setStatus('loading')
    axiosFetchServices({ filter })
      .then((list) => {
        if (cancelled) return
        setItems(list)
        setStatus('ok')
      })
      .catch(() => {
        if (cancelled) return
        const fb = fallbackServices(filter)
        setItems(fb)
        setStatus('fallback')
      })
    return () => {
      cancelled = true
    }
  }, [filter])

  return (
    <>
      <section className="search-section">
        <CatalogFilters
          filters={filters}
          onChange={setFilters}
          onImageUpload={(file) => searchByImage(file)}
          onResetImageSearch={() => resetSearch()}
          ready={ready}
        />
        <CartIcon guestCount={cartCount} />
      </section>

      {status === 'fallback' && (
        <p style={{ color: '#7f8c8d', marginTop: 8 }}>
          Бэкенд недоступен — показаны mock-данные (fallback внутри fetch).
        </p>
      )}

      <section className="cards-grid">
        {visible.map((item) => (
          <IndexedTableCard
            key={item.id}
            item={item}
            imageSimilarity={imageEmbedding ? item.score : undefined}
          />
        ))}
      </section>
      {visible.length === 0 && (
        <p style={{ color: '#7f8c8d', marginTop: 16 }}>Ничего не найдено — измените фильтры.</p>
      )}
    </>
  )
}
