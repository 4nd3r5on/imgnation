package process

import (
	"fmt"

	"imgnation-backend/pkg/types"

	"github.com/google/uuid"
)

func MakeKeyManifest(id uuid.UUID, variant string) string {
	objKey := "manifest_" + id.String()
	if variant != "" {
		objKey += "_" + variant
	}
	return objKey
}

func MakeKeyFileVariant(id uuid.UUID, variant string) string {
	objKey := id.String()
	if variant != "" {
		objKey += "_" + variant
	}
	return objKey
}

func MakeKeyChunk(
	id uuid.UUID,
	variant string,
	chunkType types.ChunkType,
	// just put any number for init chunk
	chunkIdx int64,
) string {
	objKey := id.String()
	if variant != "" {
		objKey += "_" + variant
	}
	switch chunkType {
	case types.ChunkTypeSegment:
		objKey += fmt.Sprintf("_%s_%d", string(chunkType), chunkIdx)
	case types.ChunkTypeInit:
		objKey += fmt.Sprintf("_%s", string(chunkType))
	}
	return objKey
}
