package repository

import "errors"

// Repository errors for creative brief and asset operations.
var (
	// ErrBriefNotFound indicates no creative brief matches the provided identifier.
	ErrBriefNotFound = errors.New("creative brief not found")

	// ErrAssetNotFound indicates no asset matches the provided identifier.
	ErrAssetNotFound = errors.New("asset not found")
)
