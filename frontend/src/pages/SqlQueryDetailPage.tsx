import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { AppBreadcrumbs } from '../components/AppBreadcrumbs'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import {
  deleteIndexedTableSqlQueryThunk,
  fetchIndexedTableSqlQueryDetail,
  formIndexedTableSqlQueryThunk,
  removeIndexedTableSqlQueryLineThunk,
  updateIndexedTableSqlQueryDraftThunk,
  updateIndexedTableSqlQueryLineThunk,
} from '../store/slices/indexedTableSqlQuerySlice'
import { ROUTES } from '../routePaths'
import { proxiedMediaUrl } from '../utils/proxiedMediaUrl'
import type { SerializerSqlQueryItemJSON } from '../api/Api'

type RowDraft = Pick<SerializerSqlQueryItemJSON, 'quantity' | 'position' | 'selectivity'>

export function SqlQueryDetailPage() {
  const { id: rawId } = useParams()
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const { isAuthenticated } = useAppSelector((s) => s.user)
  const { detail, detailLoading, detailError, applicationMutationLoading, itemMutationLoading } =
    useAppSelector((s) => s.indexedTableSqlQuery)

  const sqlQueryId = rawId ? Number(rawId) : NaN
  const [queryTextDraft, setQueryTextDraft] = useState('')
  const [selectivityDraft, setSelectivityDraft] = useState('')
  const [rowDrafts, setRowDrafts] = useState<Record<string, RowDraft>>({})

  useEffect(() => {
    if (!isAuthenticated) {
      navigate(ROUTES.SIGN_IN, { replace: true })
    }
  }, [isAuthenticated, navigate])

  useEffect(() => {
    if (!Number.isFinite(sqlQueryId) || sqlQueryId <= 0) return
    void dispatch(fetchIndexedTableSqlQueryDetail(sqlQueryId))
  }, [dispatch, sqlQueryId])

  const app = detail?.sql_query
  const isDraft = app?.status === 'draft'
  const busy = applicationMutationLoading || detailLoading

  useEffect(() => {
    if (!detail?.sql_query) return
    setQueryTextDraft(detail.sql_query.query_text ?? '')
    setSelectivityDraft(
      detail.sql_query.selectivity != null ? String(detail.sql_query.selectivity) : '',
    )
    const next: Record<string, RowDraft> = {}
    detail.items.forEach((row) => {
      const sid = row.indexed_table_id ?? ''
      if (!sid) return
      next[sid] = {
        quantity: row.quantity ?? 1,
        position: row.position ?? 0,
        selectivity: row.selectivity ?? 0,
      }
    })
    setRowDrafts(next)
  }, [detail])

  const updateRowDraft = useCallback((indexedTableId: string, patch: Partial<RowDraft>) => {
    setRowDrafts((prev) => {
      const base = prev[indexedTableId] ?? { quantity: 1, position: 0, selectivity: 0 }
      return { ...prev, [indexedTableId]: { ...base, ...patch } }
    })
  }, [])

  const lineBusyKey = (indexedTableId: string) =>
    Boolean(itemMutationLoading[`line-${indexedTableId}-${sqlQueryId}`])
  const rmBusy = (indexedTableId: string) => Boolean(itemMutationLoading[`rm-${indexedTableId}`])

  const handleSaveApplication = () => {
    if (!Number.isFinite(sqlQueryId) || !isDraft) return
    const sel = selectivityDraft.trim() === '' ? undefined : Number(selectivityDraft)
    void dispatch(
      updateIndexedTableSqlQueryDraftThunk({
        sqlQueryId,
        body: {
          query_text: queryTextDraft || null,
          selectivity: Number.isFinite(sel as number) ? sel : null,
        },
      }),
    )
  }

  const handleSaveRow = (indexedTableId: string) => {
    if (!Number.isFinite(sqlQueryId) || !isDraft) return
    const d = rowDrafts[indexedTableId]
    if (!d) return
    void dispatch(
      updateIndexedTableSqlQueryLineThunk({
        indexedTableId,
        sqlQueryId,
        body: {
          quantity: d.quantity ?? null,
          position: d.position ?? null,
          selectivity: d.selectivity ?? null,
        },
      }),
    )
  }

  const handleRemoveRow = (indexedTableId: string) => {
    if (!Number.isFinite(sqlQueryId) || !isDraft) return
    if (!window.confirm('Убрать индекс из заявки?')) return
    void dispatch(removeIndexedTableSqlQueryLineThunk({ indexedTableId, sqlQueryId }))
  }

  const handleForm = () => {
    if (!Number.isFinite(sqlQueryId) || !isDraft) return
    void dispatch(formIndexedTableSqlQueryThunk(sqlQueryId))
  }

  const handleDelete = () => {
    if (!Number.isFinite(sqlQueryId) || !isDraft) return
    if (!window.confirm('Удалить черновик заявки?')) return
    void dispatch(deleteIndexedTableSqlQueryThunk(sqlQueryId)).then(() => {
      navigate(ROUTES.CATALOG, { replace: true })
    })
  }

  if (!isAuthenticated) return null

  if (detailLoading && !detail) {
    return (
      <p className="ui-hint ui-loading-row" style={{ padding: '2rem', textAlign: 'center' }}>
        <span className="ui-spinner ui-spinner--lg" aria-hidden="true" /> Загрузка заявки…
      </p>
    )
  }

  if (!detail || !app || !Number.isFinite(sqlQueryId)) {
    return (
      <>
        <AppBreadcrumbs items={[{ label: 'Индексы', to: ROUTES.CATALOG }, { label: 'Заявка' }]} />
        <div className="ui-error">{detailError ?? 'Заявка не найдена.'}</div>
      </>
    )
  }

  return (
    <>
      <AppBreadcrumbs
        items={[
          { label: 'Индексы', to: ROUTES.CATALOG },
          { label: `sql_query #${sqlQueryId}` },
        ]}
      />
      <div className={`application-detail ${busy ? 'application-detail--dimmed' : ''}`}>
        <section className="request-header-card">
          <h1 className="request-title">
            sql_query #{sqlQueryId} — {app.status}
          </h1>
          <div className="request-meta">
            <div>
              <strong>Создатель:</strong> {app.creator_login ?? '—'}
            </div>
            <div>
              <strong>Позиций в заявке:</strong> {detail.items.length}
            </div>
            <div>
              <strong>Время запроса:</strong> {app.result_time ?? '—'}
            </div>
            <div>
              <strong>Память:</strong> {app.result_memory ?? '—'}
            </div>
          </div>

          {isDraft ? (
            <>
              <div className="ui-field" style={{ marginTop: 16 }}>
                <label className="ui-label" htmlFor="sql-query-text">
                  Текст / описание запроса
                </label>
                <textarea
                  id="sql-query-text"
                  className="ui-textarea"
                  rows={3}
                  value={queryTextDraft}
                  onChange={(e) => setQueryTextDraft(e.target.value)}
                  disabled={busy}
                />
              </div>
              <div className="ui-field" style={{ marginTop: 12 }}>
                <label className="ui-label" htmlFor="sql-query-sel">
                  Селективность заявки
                </label>
                <input
                  id="sql-query-sel"
                  type="number"
                  step="0.01"
                  className="search-input"
                  style={{ maxWidth: 200 }}
                  value={selectivityDraft}
                  onChange={(e) => setSelectivityDraft(e.target.value)}
                  disabled={busy}
                />
              </div>
              <div className="sql-query-actions-row">
                <button type="button" className="search-btn search-btn--sm" disabled={busy} onClick={handleSaveApplication}>
                  Сохранить поля заявки
                </button>
                <button type="button" className="search-btn search-btn--sm" disabled={busy} onClick={handleForm}>
                  Подтвердить заявку
                </button>
                <button
                  type="button"
                  className="search-btn search-btn--sm search-btn--outline"
                  disabled={busy}
                  onClick={handleForm}
                  title="Тот же вызов API, что и «Подтвердить»"
                >
                  Сформировать (dev)
                </button>
                <button
                  type="button"
                  className="search-btn search-btn--sm search-btn--danger"
                  disabled={busy}
                  onClick={handleDelete}
                >
                  Удалить черновик
                </button>
              </div>
            </>
          ) : (
            <p className="ui-hint" style={{ marginTop: 12 }}>
              Редактирование недоступно вне статуса «черновик».
            </p>
          )}
        </section>

        <section className="request-services">
          <h2 className="request-section-title">Строки заявки (indexed_table в sql_query)</h2>
          <table className="request-table">
            <thead>
              <tr>
                <th>Услуга</th>
                <th>Кол-во</th>
                <th>Позиция</th>
                <th>Селективность</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {detail.items.map((row) => {
                const sid = row.indexed_table_id ?? ''
                const draft = rowDrafts[sid]
                const img = proxiedMediaUrl(row.image_url?.trim() ? row.image_url : '/placeholder-index.png')
                return (
                  <tr key={sid}>
                    <td>
                      <div className="request-table-index">
                        <img src={img} alt={row.name} className="request-service-image" />
                        <div>
                          <div className="request-service-title">{row.name}</div>
                          <div className="request-service-id">indexed_table_id: {sid}</div>
                        </div>
                      </div>
                    </td>
                    <td>
                      {isDraft ? (
                        <input
                          type="number"
                          min={1}
                          className="ui-input-table"
                          value={draft?.quantity ?? 1}
                          onChange={(e) =>
                            updateRowDraft(sid, { quantity: Number(e.target.value) || 1 })
                          }
                          disabled={busy}
                        />
                      ) : (
                        (row.quantity ?? '—').toString()
                      )}
                    </td>
                    <td>
                      {isDraft ? (
                        <input
                          type="number"
                          className="ui-input-table"
                          value={draft?.position ?? 0}
                          onChange={(e) =>
                            updateRowDraft(sid, { position: Number(e.target.value) || 0 })
                          }
                          disabled={busy}
                        />
                      ) : (
                        (row.position ?? '—').toString()
                      )}
                    </td>
                    <td>
                      {isDraft ? (
                        <input
                          type="number"
                          step="0.01"
                          className="ui-input-table"
                          style={{ maxWidth: 88 }}
                          value={draft?.selectivity ?? 0}
                          onChange={(e) =>
                            updateRowDraft(sid, { selectivity: Number(e.target.value) || 0 })
                          }
                          disabled={busy}
                        />
                      ) : (
                        (row.selectivity ?? '—').toString()
                      )}
                    </td>
                    <td>
                      {isDraft ? (
                        <div className="sql-query-actions-row" style={{ marginTop: 0 }}>
                          <button
                            type="button"
                            className="search-btn search-btn--sm search-btn--outline"
                            disabled={busy || lineBusyKey(sid)}
                            onClick={() => handleSaveRow(sid)}
                          >
                            Сохранить строку
                          </button>
                          <button
                            type="button"
                            className="search-btn search-btn--sm search-btn--danger"
                            disabled={busy || rmBusy(sid)}
                            onClick={() => handleRemoveRow(sid)}
                          >
                            Удалить
                          </button>
                        </div>
                      ) : null}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </section>
      </div>
    </>
  )
}
