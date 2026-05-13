import { createAsyncThunk, createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { api } from '../../api'
import type {
  SerializerEditSqlQueryItemJSON,
  SerializerEditSqlQueryJSON,
  SerializerSqlQueryDetailsResponse,
  SerializerSqlQueryItemJSON,
  SerializerSqlQueryJSON,
} from '../../api/Api'
import { apiErrMessage } from '../utils/apiError'
import { logoutUser } from './userSlice'

export type SqlQueryListRow = SerializerSqlQueryJSON

export interface SqlQueryDetailState {
  sql_query: SerializerSqlQueryJSON
  items: SerializerSqlQueryItemJSON[]
}

function mapDetail(data: SerializerSqlQueryDetailsResponse | null | undefined): SqlQueryDetailState | null {
  if (!data?.sql_query || typeof data.sql_query !== 'object') return null
  return {
    sql_query: data.sql_query,
    items: Array.isArray(data.items) ? data.items : [],
  }
}

function defaultListFilters() {
  const t = new Date()
  const y = t.getFullYear()
  const m = String(t.getMonth() + 1).padStart(2, '0')
  const d = String(t.getDate()).padStart(2, '0')
  const day = `${y}-${m}-${d}`
  return { fromDate: day, toDate: day, status: '', creatorLogin: '' }
}

function buildInitialState() {
  return {
    cart: null as { id?: number; count: number } | null,
    cartLoading: false,
    detail: null as SqlQueryDetailState | null,
    detailLoading: false,
    detailError: null as string | null,
    list: [] as SqlQueryListRow[],
    listLoading: false,
    listError: null as string | null,
    filters: defaultListFilters(),
    itemMutationLoading: {} as Record<string, boolean>,
    applicationMutationLoading: false,
  }
}

type CartSliceUser = { user: { isAuthenticated: boolean } }

function emptyGuestCartPayload() {
  return { id: undefined as number | undefined, count: 0 }
}

function axiosStatus(e: unknown): number | undefined {
  if (e && typeof e === 'object' && 'response' in e) {
    const r = (e as { response?: { status?: number } }).response
    return r?.status
  }
  return undefined
}

/** Корзина sql_query: для гостя без запроса к API. */
export const fetchIndexedTableSqlQueryCart = createAsyncThunk(
  'indexedTableSqlQuery/fetchCart',
  async (_, { rejectWithValue, getState }) => {
    const before = getState() as CartSliceUser
    if (!before.user.isAuthenticated) {
      return emptyGuestCartPayload()
    }
    try {
      const r = await api.indexedTableSqlQueryApplication.indexedTableSqlQueryCartList()
      const after = getState() as CartSliceUser
      if (!after.user.isAuthenticated) {
        return emptyGuestCartPayload()
      }
      const d = r.data as Record<string, unknown>
      const rawId = d.id
      const id = typeof rawId === 'number' ? rawId : undefined
      const count = typeof d.count === 'number' ? d.count : 0
      return { id, count }
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const fetchIndexedTableSqlQueryDetail = createAsyncThunk(
  'indexedTableSqlQuery/fetchDetail',
  async (sqlQueryId: number, { rejectWithValue }) => {
    try {
      const r = await api.indexedTableSqlQueryApplication.indexedTableSqlQueryDetail(sqlQueryId)
      const detail = mapDetail(r.data)
      if (!detail) return rejectWithValue('Неверный ответ сервера')
      return detail
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const addIndexedTableToSqlQueryDraftThunk = createAsyncThunk(
  'indexedTableSqlQuery/addIndexedTableLine',
  async (indexedTableId: string, { rejectWithValue, dispatch }) => {
    try {
      await api.indexedTableSqlQueryBinding.addIndexedTableToSqlQueryDraft(indexedTableId)
      await dispatch(fetchIndexedTableSqlQueryCart())
      return indexedTableId
    } catch (e) {
      if (axiosStatus(e) === 409) {
        await dispatch(fetchIndexedTableSqlQueryCart())
        return indexedTableId
      }
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const updateIndexedTableSqlQueryLineThunk = createAsyncThunk(
  'indexedTableSqlQuery/updateIndexedTableLine',
  async (
    {
      indexedTableId,
      sqlQueryId,
      body,
    }: { indexedTableId: string; sqlQueryId: number; body: SerializerEditSqlQueryItemJSON },
    { rejectWithValue, dispatch },
  ) => {
    const key = `${indexedTableId}-${sqlQueryId}`
    try {
      await api.indexedTableSqlQueryBinding.updateIndexedTableSqlQueryLine(
        sqlQueryId,
        indexedTableId,
        body,
      )
      await dispatch(fetchIndexedTableSqlQueryDetail(sqlQueryId))
      return key
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const removeIndexedTableSqlQueryLineThunk = createAsyncThunk(
  'indexedTableSqlQuery/removeIndexedTableLine',
  async (
    { indexedTableId, sqlQueryId }: { indexedTableId: string; sqlQueryId: number },
    { rejectWithValue, dispatch },
  ) => {
    try {
      await api.indexedTableSqlQueryBinding.deleteIndexedTableSqlQueryLine(sqlQueryId, indexedTableId)
      await dispatch(fetchIndexedTableSqlQueryDetail(sqlQueryId))
      await dispatch(fetchIndexedTableSqlQueryCart())
      return indexedTableId
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const updateIndexedTableSqlQueryDraftThunk = createAsyncThunk(
  'indexedTableSqlQuery/updateSqlQueryDraft',
  async (
    { sqlQueryId, body }: { sqlQueryId: number; body: SerializerEditSqlQueryJSON },
    { rejectWithValue, dispatch },
  ) => {
    try {
      await api.indexedTableSqlQueryApplication.editIndexedTableSqlQueryUpdate(sqlQueryId, body)
      await dispatch(fetchIndexedTableSqlQueryDetail(sqlQueryId))
      return true
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const formIndexedTableSqlQueryThunk = createAsyncThunk(
  'indexedTableSqlQuery/formSqlQuery',
  async (sqlQueryId: number, { rejectWithValue, dispatch }) => {
    try {
      await api.indexedTableSqlQueryApplication.formIndexedTableSqlQueryUpdate(sqlQueryId)
      await dispatch(fetchIndexedTableSqlQueryDetail(sqlQueryId))
      await dispatch(fetchIndexedTableSqlQueryCart())
      await dispatch(fetchIndexedTableSqlQueriesListThunk())
      return true
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const deleteIndexedTableSqlQueryThunk = createAsyncThunk(
  'indexedTableSqlQuery/deleteSqlQueryDraft',
  async (sqlQueryId: number, { rejectWithValue, dispatch }) => {
    try {
      await api.indexedTableSqlQueryApplication.deleteIndexedTableSqlQueryDraft(sqlQueryId)
      await dispatch(fetchIndexedTableSqlQueryCart())
      await dispatch(fetchIndexedTableSqlQueriesListThunk())
      return true
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const finishIndexedTableSqlQueryThunk = createAsyncThunk(
  'indexedTableSqlQuery/finishSqlQuery',
  async (
    { sqlQueryId, status }: { sqlQueryId: number; status: 'completed' | 'rejected' },
    { rejectWithValue, dispatch },
  ) => {
    try {
      await api.indexedTableSqlQueryApplication.finishIndexedTableSqlQueryUpdate(sqlQueryId, {
        status,
      })
      await dispatch(fetchIndexedTableSqlQueriesListThunk())
      return true
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const fetchIndexedTableSqlQueriesListThunk = createAsyncThunk(
  'indexedTableSqlQuery/fetchList',
  async (_, { getState, rejectWithValue }) => {
    try {
      const st = getState() as {
        indexedTableSqlQuery: { filters: ReturnType<typeof defaultListFilters> }
      }
      const f = st.indexedTableSqlQuery.filters
      const query: { 'formed-from'?: string; 'formed-to'?: string; status?: string } = {}
      if (f.fromDate) query['formed-from'] = f.fromDate
      if (f.toDate) query['formed-to'] = f.toDate
      if (f.status) query.status = f.status
      const r = await api.indexedTableSqlQueryApplication.allIndexedTableSqlQueriesList(query)
      return Array.isArray(r.data) ? r.data : []
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

const indexedTableSqlQuerySlice = createSlice({
  name: 'indexedTableSqlQuery',
  initialState: buildInitialState(),
  reducers: {
    clearIndexedTableSqlQueryDetailError: (state) => {
      state.detailError = null
    },
    setIndexedTableSqlQueryListFilters: (
      state,
      action: PayloadAction<Partial<ReturnType<typeof defaultListFilters>>>,
    ) => {
      state.filters = { ...state.filters, ...action.payload }
    },
    resetIndexedTableSqlQueryListFiltersToToday: (state) => {
      state.filters = defaultListFilters()
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(logoutUser.fulfilled, () => buildInitialState())
      .addCase(logoutUser.rejected, () => buildInitialState())
      .addCase(fetchIndexedTableSqlQueryCart.pending, (state) => {
        state.cartLoading = true
      })
      .addCase(fetchIndexedTableSqlQueryCart.fulfilled, (state, action) => {
        state.cartLoading = false
        if (typeof action.payload === 'object' && action.payload && 'count' in action.payload) {
          state.cart = action.payload as typeof state.cart
        }
      })
      .addCase(fetchIndexedTableSqlQueryCart.rejected, (state) => {
        state.cartLoading = false
        state.cart = { count: 0 }
      })
      .addCase(fetchIndexedTableSqlQueryDetail.pending, (state) => {
        state.detailLoading = true
        state.detailError = null
        state.detail = null
      })
      .addCase(fetchIndexedTableSqlQueryDetail.fulfilled, (state, action) => {
        state.detailLoading = false
        state.detail = action.payload
      })
      .addCase(fetchIndexedTableSqlQueryDetail.rejected, (state, action) => {
        state.detailLoading = false
        state.detailError = action.payload as string
      })
      .addCase(fetchIndexedTableSqlQueriesListThunk.pending, (state) => {
        state.listLoading = true
        state.listError = null
      })
      .addCase(fetchIndexedTableSqlQueriesListThunk.fulfilled, (state, action) => {
        state.listLoading = false
        state.list = action.payload
      })
      .addCase(fetchIndexedTableSqlQueriesListThunk.rejected, (state, action) => {
        state.listLoading = false
        state.listError = action.payload as string
      })
      .addCase(addIndexedTableToSqlQueryDraftThunk.pending, (state) => {
        state.applicationMutationLoading = true
      })
      .addCase(addIndexedTableToSqlQueryDraftThunk.fulfilled, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(addIndexedTableToSqlQueryDraftThunk.rejected, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(updateIndexedTableSqlQueryDraftThunk.pending, (state) => {
        state.applicationMutationLoading = true
      })
      .addCase(updateIndexedTableSqlQueryDraftThunk.fulfilled, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(updateIndexedTableSqlQueryDraftThunk.rejected, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(formIndexedTableSqlQueryThunk.pending, (state) => {
        state.applicationMutationLoading = true
      })
      .addCase(formIndexedTableSqlQueryThunk.fulfilled, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(formIndexedTableSqlQueryThunk.rejected, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(deleteIndexedTableSqlQueryThunk.pending, (state) => {
        state.applicationMutationLoading = true
      })
      .addCase(deleteIndexedTableSqlQueryThunk.fulfilled, (state) => {
        state.applicationMutationLoading = false
        state.detail = null
      })
      .addCase(deleteIndexedTableSqlQueryThunk.rejected, (state) => {
        state.applicationMutationLoading = false
      })
      .addCase(updateIndexedTableSqlQueryLineThunk.pending, (state, action) => {
        const k = `${action.meta.arg.indexedTableId}-${action.meta.arg.sqlQueryId}`
        state.itemMutationLoading[`line-${k}`] = true
      })
      .addCase(updateIndexedTableSqlQueryLineThunk.fulfilled, (state, action) => {
        delete state.itemMutationLoading[`line-${action.payload}`]
      })
      .addCase(updateIndexedTableSqlQueryLineThunk.rejected, (state, action) => {
        const id = action.meta?.arg
        if (id)
          delete state.itemMutationLoading[`line-${id.indexedTableId}-${id.sqlQueryId}`]
      })
      .addCase(removeIndexedTableSqlQueryLineThunk.pending, (state, action) => {
        const id = action.meta.arg.indexedTableId
        state.itemMutationLoading[`rm-${id}`] = true
      })
      .addCase(removeIndexedTableSqlQueryLineThunk.fulfilled, (state, action) => {
        const id = action.payload
        delete state.itemMutationLoading[`rm-${id}`]
      })
      .addCase(removeIndexedTableSqlQueryLineThunk.rejected, (state, action) => {
        const id = action.meta?.arg?.indexedTableId
        if (id != null) delete state.itemMutationLoading[`rm-${id}`]
      })
      .addCase(finishIndexedTableSqlQueryThunk.pending, (state, action) => {
        const id = action.meta.arg.sqlQueryId
        state.itemMutationLoading[`finish-${id}`] = true
      })
      .addCase(finishIndexedTableSqlQueryThunk.fulfilled, (state, action) => {
        const id = action.meta.arg.sqlQueryId
        delete state.itemMutationLoading[`finish-${id}`]
      })
      .addCase(finishIndexedTableSqlQueryThunk.rejected, (state, action) => {
        const id = action.meta?.arg?.sqlQueryId
        if (id != null) delete state.itemMutationLoading[`finish-${id}`]
      })
  },
})

export const {
  clearIndexedTableSqlQueryDetailError,
  setIndexedTableSqlQueryListFilters,
  resetIndexedTableSqlQueryListFiltersToToday,
} = indexedTableSqlQuerySlice.actions
export default indexedTableSqlQuerySlice.reducer
