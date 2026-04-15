import { Outlet } from 'react-router-dom'
import { AppHeader } from '../components/AppHeader'

interface AppLayoutProps {
  cartCount: number
}

export function AppLayout(props: AppLayoutProps) {
  const { cartCount } = props
  return (
    <>
      <AppHeader cartCount={cartCount} />
      <main className="page">
        <Outlet />
      </main>
    </>
  )
}
