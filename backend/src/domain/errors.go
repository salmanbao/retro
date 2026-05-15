package domain

import "errors"

// Domain errors for authentication and profile operations.
var (
	// ErrUserNotFound indicates no user matches the provided identifier.
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailAlreadyRegistered indicates the email is already in use.
	ErrEmailAlreadyRegistered = errors.New("email already registered")

	// ErrInvalidCredentials indicates login credentials are incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrEmailNotVerified indicates the user has not verified their email.
	ErrEmailNotVerified = errors.New("email not verified")

	// ErrSessionNotFound indicates no session matches the provided token.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired indicates the session has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrTokenNotFound indicates no token matches the provided identifier.
	ErrTokenNotFound = errors.New("token not found")

	// ErrTokenExpired indicates the token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenAlreadyUsed indicates the token has already been consumed.
	ErrTokenAlreadyUsed = errors.New("token already used")

	// ErrProfileNotFound indicates no profile matches the provided identifier.
	ErrProfileNotFound = errors.New("profile not found")

	// ErrProfileNotOwned indicates the profile does not belong to the user.
	ErrProfileNotOwned = errors.New("profile not owned by user")

	// ErrInvalidEmailFormat indicates the email format is invalid.
	ErrInvalidEmailFormat = errors.New("invalid email format")

	// ErrInvalidPasswordFormat indicates the password does not meet requirements.
	ErrInvalidPasswordFormat = errors.New("invalid password format")

	// ErrInvalidProfileType indicates an unknown profile type was provided.
	ErrInvalidProfileType = errors.New("invalid profile type")

	// ErrCampaignNotFound indicates no campaign matches the provided identifier.
	ErrCampaignNotFound = errors.New("campaign not found")

	// ErrUnauthorized indicates the user is not authorized for this action.
	ErrUnauthorized = errors.New("unauthorized")
)
