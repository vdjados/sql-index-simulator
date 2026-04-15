import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'

const LOGO_DEFAULT = 'http://localhost:9000/sql-index/logo.png'

interface AppHeaderProps {
  cartCount: number
}

/**
 * Шапка как в шаблонах лаб 1–3: градиент, логотип из Minio (как в services.html), ссылки на SPA-страницы.
 */
export function AppHeader(props: AppHeaderProps) {
  const { cartCount } = props
  const { pathname } = useLocation()
  const [logoOk, setLogoOk] = useState(true)

  const headerClass = pathname.startsWith('/service/') ? 'header-secondary' : 'header-main'

  return (
    <header className={headerClass}>
      <Link to="/" className="header-logo">
        {logoOk ? (
          <img
            src={LOGO_DEFAULT}
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
      <nav className="header-actions">
        <Link to="/" className="header-nav-link">
          Индексы
        </Link>
        <Link to="/sql_query/draft" className="header-nav-link">
          Текущий запрос{cartCount > 0 ? ` (${cartCount})` : ''}
        </Link>
      </nav>
    </header>
  )
}
