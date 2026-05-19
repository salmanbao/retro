package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

// T032: Unit tests for lifecycle state machine

func TestCampaignStatus_IsValidTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.CampaignStatus
		to       domain.CampaignStatus
		expected bool
	}{
		// Valid transitions from draft
		{"draft to published", domain.CampaignStatusDraft, domain.CampaignStatusPublished, true},
		{"draft to cancelled", domain.CampaignStatusDraft, domain.CampaignStatusCancelled, true},

		// Valid transitions from published
		{"published to active", domain.CampaignStatusPublished, domain.CampaignStatusActive, true},
		{"published to cancelled", domain.CampaignStatusPublished, domain.CampaignStatusCancelled, true},

		// Valid transitions from active
		{"active to paused", domain.CampaignStatusActive, domain.CampaignStatusPaused, true},
		{"active to completed", domain.CampaignStatusActive, domain.CampaignStatusCompleted, true},
		{"active to cancelled", domain.CampaignStatusActive, domain.CampaignStatusCancelled, true},

		// Valid transitions from paused
		{"paused to active", domain.CampaignStatusPaused, domain.CampaignStatusActive, true},
		{"paused to cancelled", domain.CampaignStatusPaused, domain.CampaignStatusCancelled, true},

		// Valid transitions from completed
		{"completed to cancelled", domain.CampaignStatusCompleted, domain.CampaignStatusCancelled, false},

		// Invalid transitions
		{"draft to active", domain.CampaignStatusDraft, domain.CampaignStatusActive, false},
		{"draft to paused", domain.CampaignStatusDraft, domain.CampaignStatusPaused, false},
		{"draft to completed", domain.CampaignStatusDraft, domain.CampaignStatusCompleted, false},
		{"published to draft", domain.CampaignStatusPublished, domain.CampaignStatusDraft, false},
		{"published to paused", domain.CampaignStatusPublished, domain.CampaignStatusPaused, false},
		{"active to draft", domain.CampaignStatusActive, domain.CampaignStatusDraft, false},
		{"active to published", domain.CampaignStatusActive, domain.CampaignStatusPublished, false},
		{"paused to draft", domain.CampaignStatusPaused, domain.CampaignStatusDraft, false},
		{"paused to published", domain.CampaignStatusPaused, domain.CampaignStatusPublished, false},
		{"paused to completed", domain.CampaignStatusPaused, domain.CampaignStatusCompleted, false},
		{"completed to active", domain.CampaignStatusCompleted, domain.CampaignStatusActive, false},
		{"completed to draft", domain.CampaignStatusCompleted, domain.CampaignStatusDraft, false},
		{"completed to published", domain.CampaignStatusCompleted, domain.CampaignStatusPublished, false},
		{"completed to paused", domain.CampaignStatusCompleted, domain.CampaignStatusPaused, false},
		{"cancelled to any", domain.CampaignStatusCancelled, domain.CampaignStatusActive, false},
		{"cancelled to draft", domain.CampaignStatusCancelled, domain.CampaignStatusDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.IsValidTransition(tt.to)
			assert.Equal(t, tt.expected, result,
				"transition from %s to %s: expected %v, got %v",
				tt.from, tt.to, tt.expected, result)
		})
	}
}

func TestLifecycleStateMachine_Draft(t *testing.T) {
	status := domain.CampaignStatusDraft

	t.Run("initial state is draft", func(t *testing.T) {
		assert.Equal(t, domain.CampaignStatus("draft"), status)
	})

	t.Run("can transition to published", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusPublished))
	})

	t.Run("can transition to cancelled", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCancelled))
	})

	t.Run("cannot transition to active directly", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusActive))
	})

	t.Run("cannot transition to paused", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPaused))
	})
}

func TestLifecycleStateMachine_Published(t *testing.T) {
	status := domain.CampaignStatusPublished

	t.Run("can transition to active", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusActive))
	})

	t.Run("can transition to cancelled", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCancelled))
	})

	t.Run("cannot transition to draft", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusDraft))
	})

	t.Run("cannot transition to paused", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPaused))
	})
}

func TestLifecycleStateMachine_Active(t *testing.T) {
	status := domain.CampaignStatusActive

	t.Run("can transition to paused", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusPaused))
	})

	t.Run("can transition to completed", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})

	t.Run("can transition to cancelled", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCancelled))
	})

	t.Run("cannot transition to draft", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusDraft))
	})

	t.Run("cannot transition to published", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPublished))
	})
}

func TestLifecycleStateMachine_Paused(t *testing.T) {
	status := domain.CampaignStatusPaused

	t.Run("can transition to active", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusActive))
	})

	t.Run("can transition to cancelled", func(t *testing.T) {
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCancelled))
	})

	t.Run("cannot transition to draft", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusDraft))
	})

	t.Run("cannot transition to published", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPublished))
	})

	t.Run("cannot transition to completed", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})
}

func TestLifecycleStateMachine_Completed(t *testing.T) {
	status := domain.CampaignStatusCompleted

	t.Run("cannot transition to cancelled", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusCancelled))
	})

	t.Run("cannot transition to active", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusActive))
	})

	t.Run("cannot transition to paused", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPaused))
	})

	t.Run("cannot transition to draft", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusDraft))
	})
}

func TestLifecycleStateMachine_Cancelled(t *testing.T) {
	status := domain.CampaignStatusCancelled

	t.Run("cannot transition to any other state", func(t *testing.T) {
		assert.False(t, status.IsValidTransition(domain.CampaignStatusDraft))
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPublished))
		assert.False(t, status.IsValidTransition(domain.CampaignStatusActive))
		assert.False(t, status.IsValidTransition(domain.CampaignStatusPaused))
		assert.False(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})
}

func TestLifecycleStateMachine_FullPath(t *testing.T) {
	t.Run("happy path: draft -> published -> active -> completed", func(t *testing.T) {
		status := domain.CampaignStatusDraft

		status = domain.CampaignStatusPublished
		assert.True(t, status.IsValidTransition(domain.CampaignStatusActive))
		status = domain.CampaignStatusActive
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})

	t.Run("pause/resume path: draft -> published -> active -> paused -> active -> completed", func(t *testing.T) {
		status := domain.CampaignStatusDraft
		status = domain.CampaignStatusPublished
		status = domain.CampaignStatusActive

		// Pause
		assert.True(t, status.IsValidTransition(domain.CampaignStatusPaused))
		status = domain.CampaignStatusPaused

		// Resume
		assert.True(t, status.IsValidTransition(domain.CampaignStatusActive))
		status = domain.CampaignStatusActive

		// Complete
		assert.True(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})

	t.Run("cancellation from any state except completed", func(t *testing.T) {
		states := []domain.CampaignStatus{
			domain.CampaignStatusDraft,
			domain.CampaignStatusPublished,
			domain.CampaignStatusActive,
			domain.CampaignStatusPaused,
		}

		for _, status := range states {
			assert.True(t, status.IsValidTransition(domain.CampaignStatusCancelled),
				"%s should be cancellable", status)
		}
	})
}

func TestLifecycleStateMachine_AllStatuses(t *testing.T) {
	statuses := []domain.CampaignStatus{
		domain.CampaignStatusDraft,
		domain.CampaignStatusPublished,
		domain.CampaignStatusActive,
		domain.CampaignStatusPaused,
		domain.CampaignStatusCompleted,
		domain.CampaignStatusCancelled,
	}

	t.Run("all statuses are valid enum values", func(t *testing.T) {
		for _, s := range statuses {
			assert.NotEmpty(t, string(s), "status should have string value")
		}
	})

	t.Run("status strings match expected values", func(t *testing.T) {
		assert.Equal(t, "draft", string(domain.CampaignStatusDraft))
		assert.Equal(t, "published", string(domain.CampaignStatusPublished))
		assert.Equal(t, "active", string(domain.CampaignStatusActive))
		assert.Equal(t, "paused", string(domain.CampaignStatusPaused))
		assert.Equal(t, "completed", string(domain.CampaignStatusCompleted))
		assert.Equal(t, "cancelled", string(domain.CampaignStatusCancelled))
	})
}
