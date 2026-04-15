import { useCallback, useState } from 'react'
import { Navigate, Route, BrowserRouter as Router, Routes } from 'react-router-dom'
import { AppLayout } from './layout/AppLayout'
import { CartPage } from './pages/CartPage'
import { CatalogPage } from './pages/CatalogPage'
import { ServiceDetailPage } from './pages/ServiceDetailPage'
import { getMockById } from './mock/indexedTables'

/** Корневой компонент: роутинг и локальное состояние корзины (без Context/Redux). */
export default function App() {
  const [cartItemIDs, setCartItemIDs] = useState<string[]>([])
  const onCartAdd = useCallback((id: string) => {
    setCartItemIDs((prev) => (prev.includes(id) ? prev : [...prev, id]))
  }, [])
  const onCartClear = useCallback(() => {
    setCartItemIDs([])
  }, [])
  const existingIDs = cartItemIDs.filter((id) => !!getMockById(id))
  const cartCount = existingIDs.length

  return (
    <Router>
      <Routes>
        <Route element={<AppLayout cartCount={cartCount} />}>
          <Route path="/" element={<CatalogPage cartCount={cartCount} onCartAdd={onCartAdd} />} />
          <Route path="/catalog" element={<CatalogPage cartCount={cartCount} onCartAdd={onCartAdd} />} />
          <Route
            path="/sql_query/draft"
            element={<CartPage cartItemIDs={existingIDs} onClear={onCartClear} />}
          />
          <Route path="/service/:id" element={<ServiceDetailPage cartCount={cartCount} onCartAdd={onCartAdd} />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </Router>
  )
}
