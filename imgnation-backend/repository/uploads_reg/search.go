package uploadsReg

import (
	"context"
	"fmt"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/utils"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AggregationResult represents the structure returned by your MongoDB aggregation
type AggregationResult struct {
	Data  []types.FileMetadataPublic `bson:"data" json:"data"`
	Total []TotalCount               `bson:"total" json:"total"`
}

// TotalCount represents the count structure in the aggregation result
type TotalCount struct {
	Count int64 `bson:"count" json:"count"`
}

func (r *MongoUploadsRegistry) SearchUploads(
	ctx context.Context,
	searchParams *types.SearchUploadsParams,
	paginationParams *types.PaginationParams,
	userID uuid.UUID,
	privacyFilter bool,
) (*types.UploadsSearchResults, error) {
	// Set defaults for pagination
	if paginationParams == nil {
		paginationParams = &types.PaginationParams{}
	}
	orderAsc := types.DefaultOrderAsc
	paginationParams.Page = utils.Ternar(paginationParams.Page > 0, paginationParams.Page, types.DefaultPage)
	paginationParams.PageSize = utils.Ternar(paginationParams.PageSize > 0, paginationParams.PageSize, types.DefaultPageSize)
	paginationParams.OrderAsc = utils.Ternar(paginationParams.OrderAsc != nil, paginationParams.OrderAsc, &orderAsc)

	pipeline := r.buildSearchPipeline(
		userID, searchParams, paginationParams, privacyFilter,
	)
	cursor, err := r.uploadsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search aggregation: %w", err)
	}

	var rawDbOut []bson.Raw
	if err := cursor.All(ctx, &rawDbOut); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}
	cursor.Close(ctx)

	var total int64 = 0

	var aggResult AggregationResult
	if len(rawDbOut) > 0 {
		if err := bson.Unmarshal(rawDbOut[0], &aggResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aggregation result: %w", err)
		}
		// Extract total count
		if len(aggResult.Total) > 0 {
			total = aggResult.Total[0].Count
		}
		if len(aggResult.Total) > 0 {
			total = aggResult.Total[0].Count
		}
	}

	return &types.UploadsSearchResults{
		Items:    aggResult.Data,
		Total:    total,
		Page:     paginationParams.Page,
		PageSize: paginationParams.PageSize,
		OrderAsc: *paginationParams.OrderAsc,
	}, nil
}

func (r *MongoUploadsRegistry) buildSearchPipeline(
	userID uuid.UUID,
	searchParams *types.SearchUploadsParams,
	paginationParams *types.PaginationParams,
	privacyFilter bool,
) mongo.Pipeline {
	pipeline := mongo.Pipeline{}

	if privacyFilter {
		accessFilter := buildFileReadAccessFilter(userID)
		matchStage := bson.D{{"$match", accessFilter}}
		pipeline = append(pipeline, matchStage)
	}

	// Stage 3: Apply search filters if provided
	if searchParams != nil {
		searchFilter := buildSearchFilter(searchParams)
		if len(searchFilter) > 0 {
			searchStage := bson.D{{"$match", searchFilter}}
			pipeline = append(pipeline, searchStage)
		}
	}

	facetFilter := buildFacetFilter(paginationParams)
	pipeline = append(pipeline, facetFilter)

	return pipeline
}

// Alternative implementation using elemMatch for better array handling
func buildSearchFilter(params *types.SearchUploadsParams) bson.D {
	if params == nil {
		return bson.D{}
	}

	var uploadFilters []bson.M

	// Build filters that apply to uploads array elements
	uploadFilter := bson.M{}

	if params.UserID != uuid.Nil {
		uploadFilter["user_id"] = params.UserID
	}

	if len(params.Tags) > 0 {
		uploadFilter["tags"] = bson.M{"$all": params.Tags}
	}

	if params.MinAccessLevel >= 0 { // Always valid since 0 = visible to everyone
		uploadFilter["access.level"] = bson.M{"$gte": params.MinAccessLevel}
	}

	if params.BeforeDate != nil || params.AfterDate != nil {
		dateFilter := bson.M{}

		if params.AfterDate != nil {
			dateFilter["$gte"] = *params.AfterDate
		}

		if params.BeforeDate != nil {
			dateFilter["$lte"] = *params.BeforeDate
		}

		uploadFilter["uploaded_at"] = dateFilter
	}

	// If we have upload-specific filters, use $elemMatch
	if len(uploadFilter) > 0 {
		uploadFilters = append(uploadFilters, bson.M{
			"uploads": bson.M{"$elemMatch": uploadFilter},
		})
	}

	// If no filters, return empty filter
	if len(uploadFilters) == 0 {
		return bson.D{}
	}

	// If only one filter, return it directly
	if len(uploadFilters) == 1 {
		result := bson.D{}
		for key, value := range uploadFilters[0] {
			result = append(result, bson.E{Key: key, Value: value})
		}
		return result
	}

	// Multiple filters - combine with $and
	return bson.D{
		{Key: "$and", Value: uploadFilters},
	}
}
