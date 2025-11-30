package common

import (
	"net/url"
	"strconv"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/safeblock-dev/werr"
)

func ParsePaginationParams(q url.Values) (*types.PaginationParams, error) {
	params := &types.PaginationParams{}

	// Parse Page
	if pageStr := q.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid page parameter: %s", err.Error())
		}
		if page < 1 {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "page must be at least 1")
		}
		params.Page = page
	}

	// Parse PageSize
	if pageSizeStr := q.Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid page_size parameter: %s", err.Error())
		}
		if pageSize < 1 {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "page_size must be at least 1")
		}
		params.PageSize = pageSize
	}

	// Parse OrderAsc
	if orderAscStr := q.Get("order_asc"); orderAscStr != "" {
		switch orderAscStr {
		case "asc":
			val := true
			params.OrderAsc = &val
		case "desc":
			val := false
			params.OrderAsc = &val
		default:
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid order_asc parameter: must be 'asc' or 'desc'. Got %s", orderAscStr)
		}
	}

	return params, nil
}
