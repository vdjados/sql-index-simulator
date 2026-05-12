import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { proxiedMediaUrl } from '../utils/proxiedMediaUrl'

const LOGO_DEFAULT = 'http://localhost:9000/sql-index/logo.svg'

/**
 * Шапка как в шаблонах лаб 1–3: градиент, логотип из Minio (как в services.html), ссылки на SPA-страницы.
 */
export function AppHeader() {
  const { pathname } = useLocation()
  const [logoOk, setLogoOk] = useState(true)

  const headerClass = pathname.startsWith('/service/') ? 'header-secondary' : 'header-main'

  return (
    <header className={headerClass}>
      <Link to="/" className="header-logo">
        {logoOk ? (
          <img
            src={proxiedMediaUrl(LOGO_DEFAULT)}
            alt="SQL Index Simulator"
            onError={() => setLogoOk(false)}
          />
        ) : (
          <span className="header-title" style={{ fontSize: '1.5rem' }}>
            SQL
          </span>
        )}
        <span className="header-title">SQL Index Simulator</span>
      </Link>
    </header>
  )
}
