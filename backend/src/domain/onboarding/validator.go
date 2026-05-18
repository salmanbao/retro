package onboarding

// ProfileType constants
const (
	ProfileTypeBrand      = "brand"
	ProfileTypeEditor     = "editor"
	ProfileTypeInfluencer = "influencer"
)

// ValidProfileTypes returns the valid profile type values
var ValidProfileTypes = []string{ProfileTypeBrand, ProfileTypeEditor, ProfileTypeInfluencer}

// IsValidProfileType checks if the profile type is valid
func IsValidProfileType(profileType string) bool {
	for _, valid := range ValidProfileTypes {
		if profileType == valid {
			return true
		}
	}
	return false
}

// IsValidStepStatusTransition checks if a status transition is valid
func IsValidStepStatusTransition(current, next string) bool {
	switch current {
	case StepStatusNotStarted:
		return next == StepStatusInProgress
	case StepStatusInProgress:
		return next == StepStatusCompleted || next == StepStatusSkipped
	case StepStatusCompleted:
		return false // Cannot revert
	case StepStatusSkipped:
		return false // Cannot revert
	}
	return false
}