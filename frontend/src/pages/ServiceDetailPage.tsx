import { useEffect, useState } from 'react'
import { useStore } from 'react-redux'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { AppBreadcrumbs } from '../components/AppBreadcrumbs'
import { axiosFetchServiceById } from '../modules/servicesAxios'
import { fallbackServiceById } from '../api/fallback'
import { proxiedMediaUrl } from '../utils/proxiedMediaUrl'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { addIndexedTableToSqlQueryDraftThunk } from '../store/slices/indexedTableSqlQuerySlice'
import { useBlockingUi } from '../context/BlockingUiContext'
import { ROUTES } from '../routePaths'
import type { RootState } from '../store'
import type { ApiService } from '../api/client'

const DETAIL_GIF_FALLBACK = '/placeholder-index.gif'
const IMAGE_FALLBACK = '/placeholder-index.png'

export function ServiceDetailPage() {
  const { id: rawId } = useParams()
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const store = useStore<RootState>()
  const isAuthenticated = useAppSelector((s) => s.user.isAuthenticated)
  const { blocked } = useBlockingUi()
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
    axiosFetchServiceById(id)
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

  const handleAdd = () => {
    if (!isAuthenticated || !id || blocked) return
    void dispatch(addIndexedTableToSqlQueryDraftThunk(id)).then((a) => {
      if (addIndexedTableToSqlQueryDraftThunk.fulfilled.match(a)) {
        const draftId = store.getState().indexedTableSqlQuery.cart?.id
        if (draftId != null) navigate(ROUTES.sqlQueryDetail(draftId))
        else navigate(ROUTES.CATALOG)
      }
    })
  }

  if (!item) {
    return (
      <>
        <AppBreadcrumbs
          items={[
            { label: 'Индексы', to: ROUTES.CATALOG },
            { label: 'Не найдено' },
          ]}
        />
        <p>{status === 'loading' ? 'Загрузка…' : 'Услуга не найдена.'}</p>
        <button type="button" className="search-btn" onClick={() => navigate(ROUTES.CATALOG)}>
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
          { label: 'Индексы', to: ROUTES.CATALOG },
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
            <div className="detail-card__form">
              {isAuthenticated ? (
                <button type="button" className="search-btn" disabled={blocked} onClick={handleAdd}>
                  Добавить в новую sql_query
                </button>
              ) : (
                <p className="ui-hint">
                  <Link to={ROUTES.SIGN_IN}>Войдите</Link>, чтобы добавить услугу в заявку.
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
