import { Link } from 'react-router-dom'
import { useAppSelector } from '../store/hooks'
import { ROUTES } from '../routePaths'

interface CartIconProps {
  /** Для гостя: GET /api/cart (без JWT). */
  guestCount: number
}

/** Бейдж корзины: авторизованный пользователь — данные из Redux (GET /sql-queries/cart). */
export function CartIcon(props: CartIconProps) {
  const { guestCount } = props
  const isAuthenticated = useAppSelector((s) => s.user.isAuthenticated)
  const cart = useAppSelector((s) => s.indexedTableSqlQuery.cart)
  const count = isAuthenticated ? (cart?.count ?? 0) : guestCount
  const isEmpty = count <= 0
  const draftId = cart?.id
  const to =
    isAuthenticated && draftId != null ? ROUTES.sqlQueryDetail(draftId) : ROUTES.CATALOG

  if (isEmpty) {
    return (
      <span className="cart-icon cart-icon--empty" aria-label="Корзина пуста">
        <span className="cart-icon-symbol">🧺</span>
        <span className="cart-badge cart-badge--empty">0</span>
      </span>
    )
  }
  return (
    <Link to={to} className="cart-icon" aria-label="Корзина sql_query">
      <span className="cart-icon-symbol">🧺</span>
      <span className="cart-badge">{count}</span>
    </Link>
  )
}
