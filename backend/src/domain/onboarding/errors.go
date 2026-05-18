package onboarding

import "errors"

// Domain errors for onboarding module
var (
	// ErrStepNotFound is returned when a step is not found
	ErrStepNotFound = errors.New("onboarding step not found")

	// ErrProgressNotFound is returned when onboarding progress is not found
	ErrProgressNotFound = errors.New("onboarding progress not found")

	// ErrTemplateNotFound is returned when a template is not found
	ErrTemplateNotFound = errors.New("onboarding template not found")

	// ErrInvalidStepStatus is returned when an invalid step status transition is attempted
	ErrInvalidStepStatus = errors.New("invalid step status transition")

	// ErrStepNotSkippable is returned when attempting to skip a required step
	ErrStepNotSkippable = errors.New("cannot skip required step")

	// ErrInvalidActivationStatus is returned when an invalid activation status transition is attempted
	ErrInvalidActivationStatus = errors.New("invalid activation status transition")

	// ErrProfileNotFound is returned when a profile is not found
	ErrProfileNotFound = errors.New("profile not found")

	// ErrUnauthorized is returned when a user is not authorized to access a resource
	ErrUnauthorized = errors.New("unauthorized access to onboarding resource")

	// ErrAdminOnly is returned when a non-admin user attempts an admin-only action
	ErrAdminOnly = errors.New("admin authorization required")

	// ErrInvalidProfileType is returned when an invalid profile type is provided
	ErrInvalidProfileType = errors.New("invalid profile type for onboarding")

	// ErrProfileAlreadyOnboarded is returned when a profile is already activated
	ErrProfileAlreadyOnboarded = errors.New("profile is already activated")

	// ErrProfileNotPendingReview is returned when a profile is not in pending_review state
	ErrProfileNotPendingReview = errors.New("profile is not in pending_review state")
)