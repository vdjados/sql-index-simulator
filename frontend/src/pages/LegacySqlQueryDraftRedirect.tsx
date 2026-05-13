import { Navigate } from 'react-router-dom'
import { useAppSelector } from '../store/hooks'
import { ROUTES } from '../routePaths'

/** Старый URL лаб.6 → тематический маршрут sql_query по id черновика. */
export function LegacySqlQueryDraftRedirect() {
  const cart = useAppSelector((s) => s.indexedTableSqlQuery.cart)
  const isAuthenticated = useAppSelector((s) => s.user.isAuthenticated)
  if (!isAuthenticated) {
    return <Navigate to={ROUTES.SIGN_IN} replace />
  }
  if (cart?.id != null) {
    return <Navigate to={ROUTES.sqlQueryDetail(cart.id)} replace />
  }
  return <Navigate to={ROUTES.CATALOG} replace />
}
