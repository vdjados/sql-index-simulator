package serializer

type ErrorResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type SqlQueriesListResponse struct {
	Total int            `json:"total"`
	Items []SqlQueryJSON `json:"items"`
}

type SqlQueryDetailsResponse struct {
	SqlQuery SqlQueryJSON       `json:"sql_query"`
	Items    []SqlQueryItemJSON `json:"items"`
}

type EditSqlQueryJSON struct {
	QueryDescription *string  `json:"query_description"`
	QueryText        *string  `json:"query_text"`
	Theme            *string  `json:"theme"`
	Selectivity      *float64 `json:"selectivity"`
}

type StatusJSON struct {
	Status string `json:"status"`
}

type EditSqlQueryItemJSON struct {
	Quantity    *int     `json:"quantity"`
	Position    *int     `json:"position"`
	Selectivity *float64 `json:"selectivity"`
}
