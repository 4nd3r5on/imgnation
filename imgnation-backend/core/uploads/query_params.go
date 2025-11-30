package uploads

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

func ParseSearchUploadsParams(q url.Values) (*types.SearchUploadsParams, error) {
	params := &types.SearchUploadsParams{}

	// Parse UserID
	if userIDStr := q.Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid user_id parameter: %s", err.Error())
		}
		params.UserID = userID
	}

	// Parse Tags
	var tags []string
	for _, tagVal := range q["tags"] {
		fmt.Println(tagVal)
		parts := strings.SplitSeq(tagVal, " ")
		for part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}
	params.Tags = tags

	// Parse Public
	if minAccessLevelParam := q.Get("min_access_level"); minAccessLevelParam != "" {
		minAccessLevel, err := strconv.ParseInt(minAccessLevelParam, 10, 8)
		if err != nil {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid public parameter: %s", err.Error())
		}
		params.MinAccessLevel = types.AccessLevel(minAccessLevel)
	}

	// TODO:
	// Parse OrderBy
	// if orderByStr := q.Get("order_by"); orderByStr != "" {
	// 	if _, allowed := uploadRegistry.SearchUploadsAllowedOrderBy[orderByStr]; !allowed {
	// 		return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid order_by parameter: %s", orderByStr)
	// 	}
	// 	params.OrderBy = uploadRegistry.OrderBy(orderByStr)
	// }

	// Parse BeforeDate
	if beforeDateStr := q.Get("before_date"); beforeDateStr != "" {
		beforeDate, err := time.Parse(time.RFC3339, beforeDateStr)
		if err != nil {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid before_date parameter: %s", err.Error())
		}
		params.BeforeDate = &beforeDate
	}

	// Parse AfterDate
	if afterDateStr := q.Get("after_date"); afterDateStr != "" {
		afterDate, err := time.Parse(time.RFC3339, afterDateStr)
		if err != nil {
			return nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid after_date parameter: %s", err.Error())
		}
		params.AfterDate = &afterDate
	}

	return params, nil
}
