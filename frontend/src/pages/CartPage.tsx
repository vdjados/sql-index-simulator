import { AppBreadcrumbs } from '../components/AppBreadcrumbs'
import { getMockById } from '../mock/indexedTables'
import type { MockIndexedTable } from '../mock/indexedTables'

interface CartPageProps {
  cartItemIDs: string[]
  onClear: () => void
}

function parseSpeedMultiplier(speed: string): number {
  const match = speed.match(/x(\d+)|×(\d+)/i)
  if (!match) return 1
  const raw = match[1] ?? match[2]
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : 1
}

function parseTableSizeToMb(tableSize: string): number {
  const normalized = tableSize.trim().toUpperCase()
  const val = Number(normalized.replace(/[^\d.]/g, ''))
  if (!Number.isFinite(val) || val <= 0) return 128
  if (normalized.includes('GB')) return val * 1024
  return val
}

function estimateMetrics(item: MockIndexedTable): { timeMs: number; memoryMb: number } {
  // Легковесная front-only оценка для показа таблицы до серверного расчета.
  const k = parseSpeedMultiplier(item.speed)
  const sizeMb = parseTableSizeToMb(item.tableSize)
  const selectivity = Math.max(0.01, item.selectivity)
  const timeMs = (sizeMb * selectivity) / (k * 2.4)
  const memoryMb = sizeMb * (0.025 + selectivity*0.18)
  return { timeMs, memoryMb }
}

export function CartPage(props: CartPageProps) {
  const { cartItemIDs, onClear } = props
  const items = cartItemIDs
    .map((id) => getMockById(id))
    .filter((v): v is MockIndexedTable => v != null)
  const isEmpty = items.length === 0

  return (
    <>
      <AppBreadcrumbs
        items={[
          { label: 'Индексы', to: '/' },
          { label: 'Корзина' },
        ]}
      />
      <div className="application-detail">
        <section className="request-header-card">
          <h1 className="request-title">sql_query — черновик</h1>
          <div className="request-meta">
            <div>
              <strong>Описание запроса:</strong> —
            </div>
            <div>
              <strong>Позиции:</strong> {items.length}
            </div>
            <div>
              <strong>Время запроса:</strong> —
            </div>
            <div>
              <strong>Память запроса:</strong> —
            </div>
          </div>
          {!isEmpty && (
            <div style={{ marginTop: 16 }}>
              <button type="button" className="search-btn" onClick={onClear}>
                Очистить черновик
              </button>
            </div>
          )}
        </section>

        {isEmpty ? (
          <p style={{ color: '#526170' }}>Корзина пока пустая. Добавь индекс из каталога.</p>
        ) : (
          <section className="request-services">
            <table className="request-table">
              <thead>
                <tr>
                  <th>Услуга (Таблица + Индекс)</th>
                  <th>м-м (селективность)</th>
                  <th>Время (мс)</th>
                  <th>Память (MB)</th>
                  <th>Размер таблицы</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => {
                  const src = item.imageUrl.trim() ? item.imageUrl : '/placeholder-index.png'
                  const m = estimateMetrics(item)
                  return (
                    <tr key={item.id}>
                      <td>
                        <div className="request-table-index">
                          <img src={src} alt={item.name} className="request-service-image" />
                          <div>
                            <div className="request-service-title">{item.name}</div>
                            <div className="request-service-id">ID: {item.id}</div>
                          </div>
                        </div>
                      </td>
                      <td>{item.selectivity.toFixed(2)}</td>
                      <td>{m.timeMs.toFixed(2)}</td>
                      <td>{m.memoryMb.toFixed(2)}</td>
                      <td>{item.tableSize}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </section>
        )}
      </div>
    </>
  )
}
