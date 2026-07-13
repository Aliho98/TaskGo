package pagination

import (
	"net/url"
	"strconv"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Pagination struct {
	Page     int
	PageSize int
}

func (p Pagination) Limit() int  { return p.PageSize }
func (p Pagination) Offset() int { return (p.Page - 1) * p.PageSize }

func FromQuery(q *url.Values) Pagination {
	page, err := strconv.Atoi(q.Get("page"))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	pageSize, err := strconv.Atoi(q.Get("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Pagination{page, pageSize}
}

type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func NewMeta(p Pagination, total int64) Meta {
	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}
	return Meta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}
