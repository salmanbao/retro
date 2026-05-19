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

	// ErrAccountLocked indicates the account is locked due to failed login attempts.
	ErrAccountLocked = errors.New("account is locked due to too many failed login attempts")

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

	// ErrBriefNotFound indicates no creative brief matches the provided identifier.
	ErrBriefNotFound = errors.New("creative brief not found")

	// ErrAssetNotFound indicates no asset matches the provided identifier.
	ErrAssetNotFound = errors.New("asset not found")

	// ErrUnauthorized indicates the user is not authorized for this action.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrCSRFInvalid indicates the CSRF token is invalid or missing.
	ErrCSRFInvalid = errors.New("invalid or missing CSRF token")

	// ErrPermissionDenied indicates the caller does not have the required permission.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRoleNotFound indicates no role matches the provided identifier.
	ErrRoleNotFound = errors.New("role not found")

	// ErrPermissionNotFound indicates no permission matches the provided key.
	ErrPermissionNotFound = errors.New("permission not found")

	// ErrCircularInheritance indicates assigning this parent would create a cycle.
	ErrCircularInheritance = errors.New("circular role inheritance detected")

	// ErrMaxRolesExceeded indicates the profile has reached max role assignment.
	ErrMaxRolesExceeded = errors.New("maximum roles per profile exceeded")

	// ErrRoleAlreadyAssigned indicates this role is already assigned to the profile.
	ErrRoleAlreadyAssigned = errors.New("role already assigned to profile")

	// ErrRoleNotAssigned indicates this role is not assigned to the profile.
	ErrRoleNotAssigned = errors.New("role not assigned to profile")

	// ErrRoleHierarchyDepthExceeded indicates the parent role depth would exceed limit.
	ErrRoleHierarchyDepthExceeded = errors.New("role hierarchy depth exceeds maximum of 3")

	// ErrWildcardNotSupported indicates wildcard permissions are not yet supported.
	ErrWildcardNotSupported = errors.New("wildcard permissions not yet implemented")

	// Profile Enrichment errors
	ErrInvalidLanguageCode       = errors.New("invalid ISO 639-1 language code")
	ErrInvalidTimezone           = errors.New("invalid IANA timezone identifier")
	ErrInvalidCountryCode        = errors.New("invalid ISO 3166-1 alpha-2 country code")
	ErrInvalidCurrencyCode       = errors.New("invalid ISO 4217 currency code")
	ErrInvalidSocialLinks        = errors.New("invalid social links format")
	ErrSocialLinksTooLarge       = errors.New("social links exceeds 5KB limit")
	ErrDemographicsTooLarge      = errors.New("audience demographics exceeds 10KB limit")
	ErrEvidenceURLsTooLarge      = errors.New("evidence URLs exceed 50KB limit")
	ErrMaxPortfolioItems         = errors.New("maximum 50 portfolio items per profile")
	ErrPortfolioItemNotFound     = errors.New("portfolio item not found")
	ErrPortfolioNotEditor        = errors.New("portfolio operations require Editor profile type")
	ErrAudienceNotInfluencer     = errors.New("audience data requires Influencer profile type")
	ErrVerificationNotInfluencer = errors.New("verification requires Influencer profile type")
	ErrPayoutNotOwner            = errors.New("payout preferences can only be accessed by owner")
	ErrKYCUpdateNotAdmin         = errors.New("KYC status can only be updated by admin")
)
