package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/service"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			title:    "My Campaign",
			expected: "my-campaign",
		},
		{
			name:     "title with special characters",
			title:    "Campaign 2024! @#$-special",
			expected: "campaign-2024-special",
		},
		{
			name:     "title with multiple spaces",
			title:    "Big   Campaign   Name",
			expected: "big-campaign-name",
		},
		{
			name:     "title with hyphens preserved",
			title:    "brand-campaign-2024",
			expected: "brand-campaign-2024",
		},
		{
			name:     "title with underscores",
			title:    "quick_brown_fox",
			expected: "quick-brown-fox",
		},
		{
			name:     "title with numbers",
			title:    "Campaign 123 Test",
			expected: "campaign-123-test",
		},
		{
			name:     "title with unicode",
			title:    "Café Campaign",
			expected: "café-campaign",
		},
		{
			name:     "title with leading trailing spaces",
			title:    "  Trimmed Campaign  ",
			expected: "trimmed-campaign",
		},
		{
			name:     "title with only special chars",
			title:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "empty title",
			title:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GenerateSlug(tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlug_RemovesConsecutiveHyphens(t *testing.T) {
	result := service.GenerateSlug("Campaign---with---dashes")
	assert.Equal(t, "campaign-with-dashes", result)
}

func TestGenerateSlug_TrimsHyphensFromEnds(t *testing.T) {
	result := service.GenerateSlug("-Campaign-")
	assert.Equal(t, "campaign", result)
}

func TestValidateTimeline(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		submissionStart    time.Time
		submissionDeadline time.Time
		distributionStart  time.Time
		campaignEnd        time.Time
		expectError        bool
	}{
		{
			name:        "valid timeline",
			submissionStart:    now.Add(24 * time.Hour),
			submissionDeadline: now.Add(48 * time.Hour),
			distributionStart:  now.Add(72 * time.Hour),
			campaignEnd:        now.Add(168 * time.Hour),
			expectError:        false,
		},
		{
			name:        "submission deadline before start",
			submissionStart:    now.Add(48 * time.Hour),
			submissionDeadline: now.Add(24 * time.Hour),
			distributionStart:  now.Add(72 * time.Hour),
			campaignEnd:        now.Add(168 * time.Hour),
			expectError:        true,
		},
		{
			name:        "distribution start before deadline",
			submissionStart:    now.Add(24 * time.Hour),
			submissionDeadline: now.Add(48 * time.Hour),
			distributionStart:  now.Add(24 * time.Hour), // Same as submissionStart
			campaignEnd:        now.Add(168 * time.Hour),
			expectError:        true,
		},
		{
			name:        "campaign end before distribution start",
			submissionStart:    now.Add(24 * time.Hour),
			submissionDeadline: now.Add(48 * time.Hour),
			distributionStart:  now.Add(72 * time.Hour),
			campaignEnd:        now.Add(48 * time.Hour),
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTimeline(tt.submissionStart, tt.submissionDeadline, tt.distributionStart, tt.campaignEnd)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePayoutRange(t *testing.T) {
	tests := []struct {
		name       string
		min        *float64
		max        *float64
		expectError bool
	}{
		{
			name:        "nil values",
			min:         nil,
			max:         nil,
			expectError: false,
		},
		{
			name:        "valid range",
			min:         float64Ptr(100),
			max:         float64Ptr(200),
			expectError: false,
		},
		{
			name:        "min equals max",
			min:         float64Ptr(100),
			max:         float64Ptr(100),
			expectError: false,
		},
		{
			name:        "min greater than max",
			min:         float64Ptr(200),
			max:         float64Ptr(100),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePayoutRange(tt.min, tt.max)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}