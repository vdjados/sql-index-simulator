import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { AppBreadcrumbs } from '../components/AppBreadcrumbs'
import { fetchServices } from '../api/client'

/** Главная: проверка API в блоке в стиле карточек из старых лаб. */
export function HomePage() {
  const [apiStatus, setApiStatus] = useState<'idle' | 'loading' | 'ok' | 'error'>('idle')
  const [apiInfo, setApiInfo] = useState<string>('')

  useEffect(() => {
    let cancelled = false
    setApiStatus('loading')
    fetchServices({ filter: '' })
      .then((data) => {
        if (cancelled) return
        const n = Array.isArray(data) ? data.length : 0
        setApiInfo(`GET /api/indexed-tables: получено записей — ${n}`)
        setApiStatus('ok')
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : String(err)
        setApiInfo(msg)
        setApiStatus('error')
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <>
      <AppBreadcrumbs items={[{ label: 'Главная' }]} />
      <h1 className="page-heading" style={{ fontSize: '1.75rem', marginBottom: 12 }}>
        SQL Index Simulator
      </h1>
      <p className="lead-text-home">
        Гостевой интерфейс как в лабораторных 1–3 (те же стили <code>index_style.css</code>). Страница{' '}
        <Link to="/catalog">«Индексы»</Link> — карточки и фильтры (лаб. 5).
      </p>

      <div className="request-summary-card" style={{ maxWidth: 560, marginBottom: 20 }}>
        <div className="request-summary-header">
          <span className="request-summary-label">Состояние API</span>
          <span className="request-summary-id">бэкенд</span>
        </div>
        <div className="request-summary-body">
          {apiStatus === 'loading' && (
            <span className="ui-loading-row">
              <span className="ui-spinner" aria-hidden="true" /> Проверка…
            </span>
          )}
          {apiStatus === 'ok' && <span>{apiInfo}</span>}
          {apiStatus === 'error' && (
            <span className="ui-error" style={{ display: 'inline-block', margin: 0 }}>
              Ошибка: {apiInfo} (запустите Go на :8082 и dev-фронт на :3000)
            </span>
          )}
        </div>
      </div>

      {apiStatus === 'error' && (
        <div className="ui-warn-banner">
          Убедитесь, что включён CORS и доступен proxy <code>/api</code> → <code>localhost:8082</code>.
        </div>
      )}

      <p>
        <Link to="/catalog" className="search-btn search-btn--inline-link">
          Перейти к каталогу индексов
        </Link>
      </p>
    </>
  )
}
