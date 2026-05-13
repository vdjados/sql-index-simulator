import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'
import {
  authLoginRequest,
  authLogoutRequest,
  authRegisterRequest,
  type SignInRequest,
  type SignUpRequest,
} from '../../modules/authApi'
import { apiErrMessage } from '../utils/apiError'
import { parseIsModeratorFromToken } from '../utils/jwt'

export interface UserState {
  /** Имя и/или email для шапки */
  displayName: string
  isAuthenticated: boolean
  isModerator: boolean
  loading: boolean
  error: string | null
}

const initialState: UserState = {
  displayName: '',
  isAuthenticated: false,
  isModerator: false,
  loading: false,
  error: null,
}

export const loginUser = createAsyncThunk(
  'user/login',
  async (credentials: SignInRequest, { rejectWithValue }) => {
    try {
      await authLoginRequest(credentials)
      return { displayName: credentials.email }
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const registerUser = createAsyncThunk(
  'user/register',
  async (userData: SignUpRequest, { rejectWithValue }) => {
    try {
      await authRegisterRequest(userData)
      await authLoginRequest({ email: userData.email, password: userData.password })
      const label =
        userData.name.trim() && userData.name.trim() !== userData.email
          ? `${userData.name.trim()} (${userData.email})`
          : userData.email
      return { displayName: label }
    } catch (e) {
      return rejectWithValue(apiErrMessage(e))
    }
  },
)

export const logoutUser = createAsyncThunk('user/logout', async (_, { rejectWithValue }) => {
  try {
    await authLogoutRequest()
  } catch (e) {
    localStorage.removeItem('token')
    return rejectWithValue(apiErrMessage(e))
  }
  localStorage.removeItem('token')
  return true
})

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    clearUserError: (state) => {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginUser.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.displayName = action.payload.displayName
        const token = localStorage.getItem('token') ?? ''
        state.isModerator = parseIsModeratorFromToken(token)
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(registerUser.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(registerUser.fulfilled, (state, action) => {
        state.loading = false
        state.isAuthenticated = true
        state.displayName = action.payload.displayName
        const token = localStorage.getItem('token') ?? ''
        state.isModerator = parseIsModeratorFromToken(token)
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      .addCase(logoutUser.fulfilled, () => ({ ...initialState }))
      .addCase(logoutUser.rejected, (_state, action) => ({
        ...initialState,
        error: action.payload as string,
      }))
  },
})

export const { clearUserError } = userSlice.actions
export default userSlice.reducer
