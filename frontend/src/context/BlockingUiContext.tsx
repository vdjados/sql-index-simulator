import { createContext, useContext, useMemo, type ReactNode } from 'react'
import { useAppSelector } from '../store/hooks'

export interface BlockingUiValue {
  /** Блокировка UI при запросах к заявке / м-м (для демонстрации useContext + Redux). */
  blocked: boolean
}

const BlockingUiContext = createContext<BlockingUiValue>({ blocked: false })

export function BlockingUiProvider({ children }: { children: ReactNode }) {
  const blocked = useAppSelector((s) => {
    const q = s.indexedTableSqlQuery
    const u = s.user.loading
    return (
      u ||
      q.cartLoading ||
      q.listLoading ||
      q.detailLoading ||
      q.applicationMutationLoading ||
      Object.keys(q.itemMutationLoading).length > 0
    )
  })
  const value = useMemo(() => ({ blocked }), [blocked])
  return <BlockingUiContext.Provider value={value}>{children}</BlockingUiContext.Provider>
}

export function useBlockingUi() {
  return useContext(BlockingUiContext)
}
