import { useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { AppBreadcrumbs } from '../components/AppBreadcrumbs'
import { getMockById } from '../mock/indexedTables'

const DETAIL_GIF_FALLBACK = '/placeholder-index.gif'
const IMAGE_FALLBACK = '/placeholder-index.png'

interface ServiceDetailPageProps {
  cartCount: number
  onCartAdd: (id: string) => void
}

/** Как `service.html`: `.detail-wrapper`, `.detail-card`, бейджи таблица/скорость, кнопка `.search-btn`. */
export function ServiceDetailPage(props: ServiceDetailPageProps) {
  const { id: rawId } = useParams()
  const navigate = useNavigate()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const item = useMemo(() => (id ? getMockById(id) : undefined), [id])
  const { cartCount, onCartAdd } = props

  if (!item) {
    return (
      <>
        <AppBreadcrumbs
          items={[
            { label: 'Индексы', to: '/' },
            { label: 'Не найдено' },
          ]}
        />
        <p>Услуга не найдена в mock-данных.</p>
        <button type="button" className="search-btn" onClick={() => navigate('/')}>
          В каталог
        </button>
      </>
    )
  }

  const src = item.gifUrl.trim() ? item.gifUrl : DETAIL_GIF_FALLBACK

  return (
    <>
      <AppBreadcrumbs
        items={[
          { label: 'Индексы', to: '/' },
          { label: item.name },
        ]}
      />
      <div className="detail-wrapper">
        <div className="detail-card">
          <div className="detail-card__gif-wrap">
            <img
              src={src}
              alt={item.name}
              className="detail-card__gif"
              onError={(e) => {
                const target = e.currentTarget
                if (target.src.endsWith('/placeholder-index.png')) return
                target.src = IMAGE_FALLBACK
              }}
            />
          </div>
          <div className="detail-card__body">
            <h1 className="detail-card__title">{item.name}</h1>
            <div className="detail-card__badges">
              <span className="detail-card__badge detail-card__badge--table">Таблица: {item.tableSize}</span>
              <span className="detail-card__badge detail-card__badge--speed">Скорость: {item.speed}</span>
            </div>
            <p className="detail-card__description">{item.description}</p>
            <p style={{ fontSize: '0.85rem', color: '#7f8c8d' }}>
              Селективность: {item.selectivity.toFixed(2)} · в корзине: {cartCount}
            </p>
            <div className="detail-card__form">
              <button type="button" className="search-btn" onClick={() => onCartAdd(item.id)}>
                Добавить в запрос
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
