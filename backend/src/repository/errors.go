package repository

import "errors"

// Repository errors for creative brief, asset, and submission operations.
var (
	// ErrBriefNotFound indicates no creative brief matches the provided identifier.
	ErrBriefNotFound = errors.New("creative brief not found")

	// ErrAssetNotFound indicates no asset matches the provided identifier.
	ErrAssetNotFound = errors.New("asset not found")

	// ErrSubmissionNotFound indicates no submission matches the provided identifier.
	ErrSubmissionNotFound = errors.New("submission not found")
)
