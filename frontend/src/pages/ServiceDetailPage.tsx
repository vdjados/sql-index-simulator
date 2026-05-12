import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { AppBreadcrumbs } from '../components/AppBreadcrumbs'
import { fetchServiceById, type ApiService } from '../api/client'
import { fallbackServiceById } from '../api/fallback'
import { proxiedMediaUrl } from '../utils/proxiedMediaUrl'

const DETAIL_GIF_FALLBACK = '/placeholder-index.gif'
const IMAGE_FALLBACK = '/placeholder-index.png'

/** Как `service.html`: `.detail-wrapper`, `.detail-card`, бейджи таблица/скорость, кнопка `.search-btn`. */
export function ServiceDetailPage() {
  const { id: rawId } = useParams()
  const navigate = useNavigate()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const [item, setItem] = useState<ApiService | undefined>(undefined)
  const [status, setStatus] = useState<'idle' | 'loading' | 'ok' | 'fallback' | 'not_found'>('idle')

  useEffect(() => {
    let cancelled = false
    if (!id) {
      setStatus('not_found')
      setItem(undefined)
      return
    }
    setStatus('loading')
    fetchServiceById(id)
      .then((svc) => {
        if (cancelled) return
        setItem(svc)
        setStatus('ok')
      })
      .catch(() => {
        if (cancelled) return
        const fb = fallbackServiceById(id)
        if (!fb) {
          setStatus('not_found')
          setItem(undefined)
        } else {
          setItem(fb)
          setStatus('fallback')
        }
      })
    return () => {
      cancelled = true
    }
  }, [id])

  if (!item) {
    return (
      <>
        <AppBreadcrumbs
          items={[
            { label: 'Индексы', to: '/' },
            { label: 'Не найдено' },
          ]}
        />
        <p>{status === 'loading' ? 'Загрузка…' : 'Услуга не найдена.'}</p>
        <button type="button" className="search-btn" onClick={() => navigate('/')}>
          В каталог
        </button>
      </>
    )
  }

  const src = proxiedMediaUrl(item.video_url?.trim() ? item.video_url : DETAIL_GIF_FALLBACK)

  return (
    <>
      <AppBreadcrumbs
        items={[
          { label: 'Индексы', to: '/' },
          { label: item.name },
        ]}
      />
      {status === 'fallback' && (
        <p style={{ color: '#7f8c8d', marginTop: 8 }}>
          Бэкенд недоступен — показаны mock-данные (fallback внутри fetch).
        </p>
      )}
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
              <span className="detail-card__badge detail-card__badge--table">Таблица: {item.table_size}</span>
              <span className="detail-card__badge detail-card__badge--speed">Скорость: {item.speed}</span>
            </div>
            <p className="detail-card__description">{item.description}</p>
          </div>
        </div>
      </div>
    </>
  )
}
