import { Link } from 'react-router-dom'
import type { ApiService } from '../api/client'

const PLACEHOLDER = '/placeholder-index.png'

interface IndexedTableCardProps {
  item: ApiService
}

/** Разметка как `services.html`: .service-card, .card-image-wrapper, .service-metric-pill, .search-btn */
export function IndexedTableCard(props: IndexedTableCardProps) {
  const { item } = props
  const src = item.image_url?.trim() ? item.image_url : PLACEHOLDER

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
        <p style={{ fontSize: 13, color: '#526170', marginTop: 8 }}>{item.description}</p>
        <div style={{ marginTop: 8, textAlign: 'center' }}>
          <Link to={`/service/${encodeURIComponent(item.id)}`} style={{ fontSize: 14, fontWeight: 500 }}>
            Подробнее
          </Link>
        </div>
      </div>
    </div>
  )
}
