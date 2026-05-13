import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { loginUser } from '../store/slices/userSlice'
import { fetchIndexedTableSqlQueryCart } from '../store/slices/indexedTableSqlQuerySlice'
import { ROUTES } from '../routePaths'

export function SignInPage() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const { loading, error, isAuthenticated } = useAppSelector((s) => s.user)
  const [form, setForm] = useState({ email: '', password: '' })

  useEffect(() => {
    if (isAuthenticated) navigate(ROUTES.CATALOG, { replace: true })
  }, [isAuthenticated, navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.email || !form.password) return
    try {
      await dispatch(loginUser(form)).unwrap()
      void dispatch(fetchIndexedTableSqlQueryCart())
      navigate(ROUTES.CATALOG, { replace: true })
    } catch {
      void 0
    }
  }

  return (
    <div className="auth-page-min">
      <h1 className="page-heading" style={{ marginBottom: '1rem' }}>
        Вход
      </h1>
      {error ? <div className="ui-error">{error}</div> : null}
      <form onSubmit={handleSubmit}>
        <label className="auth-page-min__label" htmlFor="signin-email">
          Email
        </label>
        <input
          id="signin-email"
          className="search-input"
          type="email"
          value={form.email}
          onChange={(e) => setForm({ ...form, email: e.target.value })}
          required
          disabled={loading}
          autoComplete="username"
        />
        <label className="auth-page-min__label" htmlFor="signin-password">
          Пароль
        </label>
        <input
          id="signin-password"
          className="search-input"
          type="password"
          value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })}
          required
          disabled={loading}
          autoComplete="current-password"
        />
        <button type="submit" className="search-btn" style={{ width: '100%', marginTop: 16 }} disabled={loading}>
          {loading ? (
            <span className="ui-loading-row">
              <span className="ui-spinner" aria-hidden="true" /> Вход…
            </span>
          ) : (
            'Войти'
          )}
        </button>
      </form>
      <p style={{ marginTop: 16 }} className="ui-hint">
        Нет аккаунта? <Link to={ROUTES.SIGN_UP}>Регистрация</Link>
      </p>
    </div>
  )
}
