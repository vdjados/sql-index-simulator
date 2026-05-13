import { configureStore } from '@reduxjs/toolkit'
import userReducer from './slices/userSlice'
import indexedTableSqlQueryReducer from './slices/indexedTableSqlQuerySlice'

/**
 * Redux Toolkit: reducer-ы user и indexedTableSqlQuery; middleware по умолчанию включает redux-thunk.
 */
export const store = configureStore({
  reducer: {
    user: userReducer,
    indexedTableSqlQuery: indexedTableSqlQueryReducer,
  },
  devTools: true,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      thunk: true,
    }),
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
