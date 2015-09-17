package pagination

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// DefaultLimit defines the default number of items per page for APIs
	DefaultLimit int64 = 25

	// DefaultOffset defines the default offset for API responses
	DefaultOffset int64 = 0
)

// Core contains the fields that encapsulate pagination of arrays
type Core struct {
	Total     int64  `json:"total"`
	Limit     int64  `json:"limit"`
	Offset    int64  `json:"offset"`
	MaxOffset int64  `json:"maxOffset"`
	Pages     int64  `json:"totalPages"`
	Page      int64  `json:"page"`
	Type      string `json:"type"`
}

// Pagination describes an array in JSON and how to paginate the collection
// NOTE: You are advised to use your own strongly typed one, this one is only
// good for marshalling to JSON and cannot be used to marshal back to an object
// as it uses an interface{} for the Items
type Pagination struct {
	Core
	Items interface{} `json:"items"`
}

// Populate will populate the pagination fields of an array without setting
// the items
func (m *Core) Populate(
	total int64,
	limit int64,
	offset int64,
	contentType string,
) {
	m.Total = total
	m.Limit = limit
	m.Offset = offset
	m.MaxOffset = MaxOffset(total, limit)
	m.Pages = PageCount(total, limit)
	m.Page = CurrentPage(offset, limit)
	m.Type = contentType
}

// Construct returns a Pagination fully populated
func Construct(
	resources interface{},
	contentType string,
	total int64,
	limit int64,
	offset int64,
) Pagination {
	m := Pagination{}
	m.Populate(total, limit, offset, contentType)
	m.Items = resources
	return m
}

// CurrentPage returns the current page for a given offset and limit value
func CurrentPage(offset int64, limit int64) int64 {
	if limit == 0 {
		return 0
	}

	return (offset + limit) / limit
}

// LimitAndOffset returns the Limit and Offset for a given request querystring
func LimitAndOffset(query url.Values) (int64, int64, int, error) {
	var (
		limit  int64
		offset int64
	)

	limit = DefaultLimit
	limitParam := "limit"

	if query.Get("per_page") != "" {
		limitParam = "per_page"
	}

	if query.Get(limitParam) != "" {
		inLimit, err := strconv.ParseInt(query.Get(limitParam), 10, 64)
		if err != nil {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("%s (%s) is not a number", limitParam, query.Get(limitParam))
		}
		limit = inLimit
	}

	if limit != DefaultLimit {
		if limit < 1 {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("%s (%d) cannot be zero or negative", limitParam, limit)
		}

		if limit%5 != 0 {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("%s (%d) must be a multiple of 5", limitParam, limit)
		}

		const maxLimit = 250
		if limit > maxLimit {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("%s (%d) cannot exceed %d", limitParam, limit, maxLimit)
		}
	}

	offset = DefaultOffset
	if query.Get("offset") != "" {
		inOffset, err := strconv.ParseInt(query.Get("offset"), 10, 64)
		if err != nil {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("offset (%s) is not a number", query.Get("offset"))
		}

		if inOffset < 0 {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("offset (%d) cannot be negative", inOffset)
		}

		if inOffset%limit != 0 {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf(
					"offset (%d) must be a multiple of limit (%d) or zero",
					inOffset,
					limit,
				)
		}

		offset = inOffset
	}

	if offset == DefaultOffset && query.Get("page") != "" {
		inPage, err := strconv.ParseInt(query.Get("page"), 10, 64)
		if err != nil {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("page (%s) is not a number", query.Get("page"))
		}

		if inPage <= 0 {
			return 0, 0, http.StatusBadRequest,
				fmt.Errorf("page (%d) must be 1 or higher", inPage)
		}

		// Calculate offset from page
		offset = inPage*limit - limit
	}

	return limit, offset, http.StatusOK, nil
}

// MaxOffset returns the maximum possible offset for a given number of
// pages and limit per page
func MaxOffset(total int64, limit int64) int64 {
	if limit == 0 {
		limit = DefaultLimit
	}
	return ((total - 1) / limit) * limit
}

// PageCount returns the number of pages for a given total and items per
// page
func PageCount(total int64, limit int64) int64 {
	if limit == 0 {
		limit = DefaultLimit
	}

	pages := total / limit

	if total%limit > 0 {
		pages++
	}

	return pages
}

// OffsetFromPage returns the offset from a page number. This helps older
// interfaces continue to support pageNumber and perPage parameters whilst
// we would use limit and offset internally.
func OffsetFromPage(page int64, limit int64) (offset int64) {
	offset = DefaultOffset

	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = DefaultLimit
	}

	return (page * limit) - limit
}
