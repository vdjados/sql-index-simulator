import axios from 'axios'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api'

/** Axios только для /users/* (без codegen). */
export const authAxios = axios.create({
  baseURL,
})

authAxios.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

authAxios.interceptors.response.use(
  (response) => {
    const data = response.data as Record<string, unknown> | null
    if (data && typeof data === 'object' && 'token' in data && data.token) {
      localStorage.setItem('token', String(data.token))
    }
    return response
  },
  (error) => Promise.reject(error),
)

export interface SignInRequest {
  email: string
  password: string
}

export interface SignUpRequest {
  email: string
  name: string
  password: string
}

export async function authLoginRequest(credentials: SignInRequest) {
  return authAxios.post<{ token?: string; role?: string }>('/users/login', credentials)
}

export async function authRegisterRequest(user: SignUpRequest) {
  return authAxios.post('/users/register', user)
}

export async function authLogoutRequest() {
  return authAxios.post('/users/logout', null, {
    validateStatus: (s) => s === 204 || s < 300,
  })
}
