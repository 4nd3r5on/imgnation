package process

import (
	"context"

	"imgnation-backend/pkg/types"
	"imgnation-backend/repository/storage"
)

// ProcessingStrategy defines how variants should be processed
type ProcessingStrategy int

const (
	StrategySequential ProcessingStrategy = iota
	StrategyMultivariant
)

type VariantGroup[VariantOpts any] struct {
	Name     string
	Strategy ProcessingStrategy
	Variants []VariantOpts
}

type Format[VariantOpts any] struct {
	Name           string
	Format         string
	CreateVariants []VariantGroup[VariantOpts]
}

// Creates file variant, puts it in storage and returns info
type VariantFunc func(
	ctx context.Context,
	storage storage.IStorage,
) ([]types.FileVariantInfo, []types.ChunkInfo, error)

// Returns functions for creating variants
// So we can later execute them in worker pool
type ProcessFunc func(
	ctx context.Context,
	opts *types.ProcessFileOpts,
) ([]VariantFunc, error)
