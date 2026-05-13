import { Link } from 'react-router-dom'
import type { ApiService } from '../api/client'
import { proxiedMediaUrl } from '../utils/proxiedMediaUrl'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { addIndexedTableToSqlQueryDraftThunk } from '../store/slices/indexedTableSqlQuerySlice'
import { useBlockingUi } from '../context/BlockingUiContext'
import { ROUTES } from '../routePaths'

const PLACEHOLDER = '/placeholder-index.png'

interface IndexedTableCardProps {
  item: ApiService
  /** Cosine similarity vs uploaded image (browser SigLIP); set only after image search. */
  imageSimilarity?: number
}

/** Карточка indexed_table: «Добавить в sql_query» через thunk + axios (codegen API). */
export function IndexedTableCard(props: IndexedTableCardProps) {
  const { item, imageSimilarity } = props
  const raw = item.image_url?.trim() ? item.image_url : PLACEHOLDER
  const src = proxiedMediaUrl(raw)
  const dispatch = useAppDispatch()
  const isAuthenticated = useAppSelector((s) => s.user.isAuthenticated)
  const { blocked } = useBlockingUi()

  const handleAdd = () => {
    if (!isAuthenticated || blocked) return
    void dispatch(addIndexedTableToSqlQueryDraftThunk(item.id))
  }

  return (
    <div className="service-card">
      <Link to={`/service/${encodeURIComponent(item.id)}`}>
        <div className="card-image-wrapper">
          <img src={src} alt={item.name} className="service-card-image" />
        </div>
      </Link>
      <div className="service-card-body">
        <h1 className="service-card-title">{item.name}</h1>
        <div className="service-metrics">
          <span className="service-metric-pill">Таблица + Индекс: {item.table_size}</span>
        </div>
        {imageSimilarity !== undefined && (
          <p className="similarity-badge">Similarity: {(imageSimilarity * 100).toFixed(1)}%</p>
        )}
        <p className="service-card-desc">{item.description}</p>
        <div className="service-card-actions">
          <Link to={`/service/${encodeURIComponent(item.id)}`} className="card-more-link">
            Подробнее
          </Link>
          {isAuthenticated ? (
            <div className="service-add-form">
              <button type="button" className="search-btn" disabled={blocked} onClick={handleAdd}>
                Добавить в новую sql_query
              </button>
            </div>
          ) : (
            <span className="card-guest-hint">
              <Link to={ROUTES.SIGN_IN}>Войдите</Link>, чтобы добавить в заявку
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
