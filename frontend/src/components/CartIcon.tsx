import { Link } from 'react-router-dom'

interface CartIconProps {
  count: number
}

/** Как в services.html: 🧺 и бейдж; при 0 — класс cart-icon--empty. */
export function CartIcon(props: CartIconProps) {
  const { count } = props
  const isEmpty = count <= 0
  if (isEmpty) {
    return (
      <span className="cart-icon cart-icon--empty" aria-label="Корзина пуста">
        <span className="cart-icon-symbol">🧺</span>
        <span className="cart-badge cart-badge--empty">0</span>
      </span>
    )
  }
  return (
    <Link to="/sql_query/draft" className="cart-icon" aria-label="Корзина">
      <span className="cart-icon-symbol">🧺</span>
      <span className="cart-badge">{count}</span>
    </Link>
  )
}
