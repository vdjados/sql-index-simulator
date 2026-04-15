import { Link } from 'react-router-dom'
import type { MockIndexedTable } from '../mock/indexedTables'

const PLACEHOLDER = '/placeholder-index.png'

interface IndexedTableCardProps {
  item: MockIndexedTable
  onAddToCart: (id: string) => void
}

/** Разметка как `services.html`: .service-card, .card-image-wrapper, .service-metric-pill, .search-btn */
export function IndexedTableCard(props: IndexedTableCardProps) {
  const { item, onAddToCart } = props
  const src = item.imageUrl.trim() ? item.imageUrl : PLACEHOLDER

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
          <span className="service-metric-pill">Таблица + Индекс: {item.tableSize}</span>
        </div>
        <p style={{ fontSize: 13, color: '#526170', marginTop: 8 }}>{item.description}</p>
        <div className="service-add-form">
          <button type="button" className="search-btn" onClick={() => onAddToCart(item.id)}>
            Добавить в запрос
          </button>
        </div>
        <div style={{ marginTop: 8, textAlign: 'center' }}>
          <Link to={`/service/${encodeURIComponent(item.id)}`} style={{ fontSize: 14, fontWeight: 500 }}>
            Подробнее
          </Link>
        </div>
      </div>
    </div>
  )
}
