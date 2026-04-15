import { Link } from 'react-router-dom'

export interface BreadcrumbItem {
  label: string
  to?: string
}

interface AppBreadcrumbsProps {
  /** Цепочка от базовой страницы к текущей (лаба 5: самописные breadcrumbs без Redux/Context). */
  items: BreadcrumbItem[]
}

export function AppBreadcrumbs(props: AppBreadcrumbsProps) {
  const { items } = props
  return (
    <nav className="app-breadcrumbs" aria-label="Навигационная цепочка">
      <ol className="app-breadcrumbs__list">
        {items.map((item, index) => {
          const isLast = index === items.length - 1
          return (
            <li key={`${item.label}-${index}`} className="app-breadcrumbs__item">
              {item.to && !isLast ? (
                <Link to={item.to} className="app-breadcrumbs__link">
                  {item.label}
                </Link>
              ) : (
                <span className={isLast ? 'app-breadcrumbs__current' : undefined}>{item.label}</span>
              )}
              {!isLast && <span className="app-breadcrumbs__sep" aria-hidden="true">/</span>}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
