package common

type QueryOptions struct {
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
	SortBy   string                 `json:"sort_by"`
	SortDesc bool                   `json:"sort_desc"`
	Filters  map[string]interface{} `json:"filters"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total      int64 `json:"total" example:"100"`
	Limit      int   `json:"limit" example:"20"`
	Page       int   `json:"page" example:"1"`
	TotalPages int   `json:"total_pages" example:"5"`
	NextPage   *int  `json:"next_page,omitempty"`
	PrevPage   *int  `json:"prev_page,omitempty"`
}

func (q *QueryOptions) Offset() int {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	return (q.Page - 1) * q.Limit
}

func (q *QueryOptions) LimitValue() int {
	if q.Limit <= 0 {
		return 20
	}
	if q.Limit > 100 {
		return 100
	}
	return q.Limit
}

type PageResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}
