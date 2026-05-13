/** URL по теме задания: sql_query, indexed-tables. */
export const ROUTES = {
  CATALOG: '/',
  CATALOG_ALT: '/catalog',
  SIGN_IN: '/signin',
  SIGN_UP: '/signup',
  SQL_QUERIES: '/sql-queries',
  sqlQueryDetail: (id: number) => `/sql-query/${id}`,
} as const
