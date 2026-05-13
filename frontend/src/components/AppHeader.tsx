import { useState, type MouseEvent } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { proxiedMediaUrl } from '../utils/proxiedMediaUrl'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { logoutUser } from '../store/slices/userSlice'
import { ROUTES } from '../routePaths'

const LOGO_DEFAULT = 'http://localhost:9000/sql-index/logo.svg'

export function AppHeader() {
  const { pathname } = useLocation()
  const [logoOk, setLogoOk] = useState(true)
  const dispatch = useAppDispatch()
  const { isAuthenticated, displayName } = useAppSelector((s) => s.user)
  const cart = useAppSelector((s) => s.indexedTableSqlQuery.cart)

  const headerClass = pathname.startsWith('/service/') ? 'header-secondary' : 'header-main'

  const draftReady = Boolean(isAuthenticated && cart?.id != null)
  const draftTo = draftReady && cart?.id != null ? ROUTES.sqlQueryDetail(cart.id) : ROUTES.CATALOG

  const handleLogout = (e: MouseEvent) => {
    e.preventDefault()
    void dispatch(logoutUser())
  }

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

      <div className="header-actions">
        <Link to={ROUTES.CATALOG} className="header-nav-link">
          Индексы (каталог)
        </Link>
        {isAuthenticated ? (
          <Link to={ROUTES.SQL_QUERIES} className="header-nav-link">
            Заявки sql_query
          </Link>
        ) : null}
        {draftReady ? (
          <Link to={draftTo} className="header-nav-link">
            Текущая sql_query (черновик)
          </Link>
        ) : (
          <span className="header-nav-link header-nav-link--muted">Текущая sql_query (черновик)</span>
        )}
        {isAuthenticated ? <span className="header-user-label">{displayName}</span> : null}
        {isAuthenticated ? (
          <a href="/" className="header-nav-link" onClick={handleLogout}>
            Выход
          </a>
        ) : (
          <>
            <Link to={ROUTES.SIGN_IN} className="header-nav-link">
              Вход
            </Link>
            <Link to={ROUTES.SIGN_UP} className="header-nav-link">
              Регистрация
            </Link>
          </>
        )}
      </div>
    </header>
  )
}
