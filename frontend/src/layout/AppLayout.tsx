import { Outlet } from 'react-router-dom'
import { useEffect } from 'react'
import { BlockingUiProvider, useBlockingUi } from '../context/BlockingUiContext'
import { AppHeader } from '../components/AppHeader'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { fetchIndexedTableSqlQueryCart } from '../store/slices/indexedTableSqlQuerySlice'

function GlobalBlockingOverlay() {
  const { blocked } = useBlockingUi()
  if (!blocked) return null
  return (
    <div className="global-blocking-overlay" aria-live="polite" aria-busy="true">
      <span className="ui-spinner ui-spinner--lg" aria-hidden="true" />
      <span className="global-blocking-overlay__text">Запрос…</span>
    </div>
  )
}

function AppLayoutInner() {
  const dispatch = useAppDispatch()
  const isAuthenticated = useAppSelector((s) => s.user.isAuthenticated)

  useEffect(() => {
    void dispatch(fetchIndexedTableSqlQueryCart())
  }, [dispatch, isAuthenticated])

  return (
    <>
      <GlobalBlockingOverlay />
      <AppHeader />
      <main className="page page--with-overlay-host">
        <Outlet />
      </main>
    </>
  )
}

export function AppLayout() {
  return (
    <BlockingUiProvider>
      <AppLayoutInner />
    </BlockingUiProvider>
  )
}
