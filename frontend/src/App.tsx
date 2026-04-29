import { useState } from 'react'
import { Navigate, Route, BrowserRouter as Router, Routes } from 'react-router-dom'
import { AppLayout } from './layout/AppLayout'
import { CartPage } from './pages/CartPage'
import { CatalogPage } from './pages/CatalogPage'
import { ServiceDetailPage } from './pages/ServiceDetailPage'
import { getMockById } from './mock/indexedTables'

/** Корневой компонент: роутинг и локальное состояние корзины (без Context/Redux). */
export default function App() {
  const [cartItemIDs, setCartItemIDs] = useState<string[]>([])

  const onCartClear = () => {
    setCartItemIDs([])
  }
  const existingIDs = cartItemIDs.filter((id) => !!getMockById(id))

  return (
    <Router>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<CatalogPage />} />
          <Route path="/catalog" element={<CatalogPage />} />
          <Route
            path="/sql_query/draft"
            element={<CartPage cartItemIDs={existingIDs} onClear={onCartClear} />}
          />
          <Route path="/service/:id" element={<ServiceDetailPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </Router>
  )
}
