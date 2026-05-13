/**
 * Клиент в стиле codegen из Swagger 2.0 (axios): домены sql-queries и sql-query-items.
 * Услуги (indexed-tables) и users — отдельные axios-модули.
 */

import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  HeadersDefaults,
  ResponseType,
} from 'axios'
import axios from 'axios'

/** serializer.SqlQueryJSON */
export interface SerializerSqlQueryJSON {
  id?: number
  status?: string
  created_at?: string
  created_by_id?: number
  formed_at?: string | null
  completed_at?: string | null
  moderator_id?: number | null
  creator_login?: string
  moderator_login?: string | null
  query_text?: string
  theme?: string
  selectivity?: number
  result_time?: string
  result_memory?: string
  results_count?: number
}

/** serializer.SqlQueryItemJSON */
export interface SerializerSqlQueryItemJSON {
  indexed_table_id?: string
  name?: string
  table_size?: string
  selectivity?: number
  image_url?: string
  video_url?: string
  quantity?: number
  position?: number
  calculated_time_ms?: number | null
}

/** serializer.SqlQueryDetailsResponse */
export interface SerializerSqlQueryDetailsResponse {
  sql_query?: SerializerSqlQueryJSON
  items?: SerializerSqlQueryItemJSON[]
}

/** serializer.EditSqlQueryJSON */
export interface SerializerEditSqlQueryJSON {
  query_description?: string | null
  query_text?: string | null
  theme?: string | null
  selectivity?: number | null
}

/** serializer.EditSqlQueryItemJSON */
export interface SerializerEditSqlQueryItemJSON {
  quantity?: number | null
  position?: number | null
  selectivity?: number | null
}

/** serializer.StatusJSON */
export interface SerializerStatusJSON {
  status?: string
}

export type QueryParamsType = Record<string | number, unknown>

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, 'data' | 'params' | 'url' | 'responseType'> {
  secure?: boolean
  path: string
  type?: ContentType
  query?: QueryParamsType
  format?: ResponseType
  body?: unknown
}

export type RequestParams = Omit<FullRequestParams, 'body' | 'method' | 'query' | 'path'>

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, 'data' | 'cancelToken'> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void
  secure?: boolean
  format?: ResponseType
}

export const ContentType = {
  Json: 'application/json',
  JsonApi: 'application/vnd.api+json',
  FormData: 'multipart/form-data',
  UrlEncoded: 'application/x-www-form-urlencoded',
  Text: 'text/plain',
} as const

export type ContentType = (typeof ContentType)[keyof typeof ContentType]

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance
  private securityData: SecurityDataType | null = null
  private securityWorker?: ApiConfig<SecurityDataType>['securityWorker']
  private secure?: boolean
  private format?: ResponseType

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || '/api',
    })
    this.secure = secure
    this.format = format
    this.securityWorker = securityWorker
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data
  }

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method)

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[method.toLowerCase() as keyof HeadersDefaults]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    }
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === 'object' && formItem !== null) {
      return JSON.stringify(formItem)
    }
    return `${formItem}`
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key]
      const propertyContent: unknown[] = property instanceof Array ? property : [property]

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File
        formData.append(key, isFileType ? formItem : this.stringifyFormItem(formItem))
      }

      return formData
    }, new FormData())
  }

  public request = async <T = unknown>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === 'boolean' ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {}
    const requestParams = this.mergeRequestParams(params, secureParams)
    const responseFormat = format || this.format || undefined

    let reqBody: unknown = body
    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === 'object'
    ) {
      reqBody = this.createFormData(body as Record<string, unknown>)
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== 'string'
    ) {
      reqBody = JSON.stringify(body)
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type ? { 'Content-Type': type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: reqBody,
      url: path,
    })
  }
}

export class Api<SecurityDataType = unknown> extends HttpClient<SecurityDataType> {
  /** Связь indexed_table ↔ sql_query (м-м через строки заявки). */
  indexedTableSqlQueryBinding = {
    addIndexedTableToSqlQueryDraft: (indexedTableId: string, params: RequestParams = {}) =>
      this.request<SerializerSqlQueryItemJSON>({
        path: `/sql-queries/draft/indexed-tables/${encodeURIComponent(indexedTableId)}`,
        method: 'POST',
        secure: true,
        format: 'json',
        ...params,
      }),

    updateIndexedTableSqlQueryLine: (
      sqlQueryId: number,
      indexedTableId: string,
      data: SerializerEditSqlQueryItemJSON,
      params: RequestParams = {},
    ) =>
      this.request<SerializerSqlQueryItemJSON>({
        path: `/sql-queries/${sqlQueryId}/indexed-tables/${encodeURIComponent(indexedTableId)}`,
        method: 'PUT',
        body: data,
        secure: true,
        type: ContentType.Json,
        format: 'json',
        ...params,
      }),

    deleteIndexedTableSqlQueryLine: (
      sqlQueryId: number,
      indexedTableId: string,
      params: RequestParams = {},
    ) =>
      this.request<Record<string, unknown>>({
        path: `/sql-queries/${sqlQueryId}/indexed-tables/${encodeURIComponent(indexedTableId)}`,
        method: 'DELETE',
        secure: true,
        format: 'json',
        ...params,
      }),
  }

  /** Заявка sql_query */
  indexedTableSqlQueryApplication = {
    indexedTableSqlQueryCartList: (params: RequestParams = {}) =>
      this.request<Record<string, unknown>>({
        path: `/sql-queries/cart`,
        method: 'GET',
        format: 'json',
        ...params,
      }),

    allIndexedTableSqlQueriesList: (
      query?: {
        status?: string
        'formed-from'?: string
        'formed-to'?: string
        'from-date'?: string
        'to-date'?: string
      },
      params: RequestParams = {},
    ) =>
      this.request<SerializerSqlQueryJSON[]>({
        path: `/sql-queries`,
        method: 'GET',
        query,
        secure: true,
        format: 'json',
        ...params,
      }),

    indexedTableSqlQueryDetail: (id: number, params: RequestParams = {}) =>
      this.request<SerializerSqlQueryDetailsResponse>({
        path: `/sql-queries/${id}`,
        method: 'GET',
        secure: true,
        format: 'json',
        ...params,
      }),

    editIndexedTableSqlQueryUpdate: (
      id: number,
      body: SerializerEditSqlQueryJSON,
      params: RequestParams = {},
    ) =>
      this.request<SerializerSqlQueryJSON>({
        path: `/sql-queries/${id}`,
        method: 'PUT',
        body,
        secure: true,
        type: ContentType.Json,
        format: 'json',
        ...params,
      }),

    deleteIndexedTableSqlQueryDraft: (id: number, params: RequestParams = {}) =>
      this.request<Record<string, unknown>>({
        path: `/sql-queries/${id}`,
        method: 'DELETE',
        secure: true,
        format: 'json',
        ...params,
      }),

    formIndexedTableSqlQueryUpdate: (id: number, params: RequestParams = {}) =>
      this.request<SerializerSqlQueryJSON>({
        path: `/sql-queries/${id}/form`,
        method: 'PUT',
        secure: true,
        format: 'json',
        ...params,
      }),

    finishIndexedTableSqlQueryUpdate: (
      id: number,
      status: SerializerStatusJSON,
      params: RequestParams = {},
    ) =>
      this.request<SerializerSqlQueryJSON>({
        path: `/sql-queries/${id}/finish`,
        method: 'PUT',
        body: status,
        secure: true,
        type: ContentType.Json,
        format: 'json',
        ...params,
      }),
  }
}
