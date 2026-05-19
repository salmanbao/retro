package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
)

// T033: Integration tests for all lifecycle transitions

func TestCampaignLifecycle_AllTransitions(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("draft can transition to published", func(t *testing.T) {
		status := domain.CampaignStatusDraft
		require.True(t, status.IsValidTransition(domain.CampaignStatusPublished),
			"draft should transition to published")
	})

	t.Run("draft can transition to cancelled", func(t *testing.T) {
		status := domain.CampaignStatusDraft
		require.True(t, status.IsValidTransition(domain.CampaignStatusCancelled),
			"draft should transition to cancelled")
	})

	t.Run("published can transition to active", func(t *testing.T) {
		status := domain.CampaignStatusPublished
		require.True(t, status.IsValidTransition(domain.CampaignStatusActive),
			"published should transition to active")
	})

	t.Run("active can transition to paused", func(t *testing.T) {
		status := domain.CampaignStatusActive
		require.True(t, status.IsValidTransition(domain.CampaignStatusPaused),
			"active should transition to paused")
	})

	t.Run("active can transition to completed", func(t *testing.T) {
		status := domain.CampaignStatusActive
		require.True(t, status.IsValidTransition(domain.CampaignStatusCompleted),
			"active should transition to completed")
	})

	t.Run("active can transition to cancelled", func(t *testing.T) {
		status := domain.CampaignStatusActive
		require.True(t, status.IsValidTransition(domain.CampaignStatusCancelled),
			"active should transition to cancelled")
	})

	t.Run("paused can transition to active", func(t *testing.T) {
		status := domain.CampaignStatusPaused
		require.True(t, status.IsValidTransition(domain.CampaignStatusActive),
			"paused should transition to active")
	})

	t.Run("paused can transition to cancelled", func(t *testing.T) {
		status := domain.CampaignStatusPaused
		require.True(t, status.IsValidTransition(domain.CampaignStatusCancelled),
			"paused should transition to cancelled")
	})

	t.Run("completed can transition to cancelled", func(t *testing.T) {
		status := domain.CampaignStatusCompleted
		require.True(t, status.IsValidTransition(domain.CampaignStatusCancelled),
			"completed should transition to cancelled")
	})
}

func TestCampaignLifecycle_InvalidTransitions(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("draft cannot skip to active", func(t *testing.T) {
		status := domain.CampaignStatusDraft
		require.False(t, status.IsValidTransition(domain.CampaignStatusActive),
			"draft should not skip to active")
	})

	t.Run("published cannot go back to draft", func(t *testing.T) {
		status := domain.CampaignStatusPublished
		require.False(t, status.IsValidTransition(domain.CampaignStatusDraft),
			"published should not go back to draft")
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		status := domain.CampaignStatusCancelled
		require.False(t, status.IsValidTransition(domain.CampaignStatusDraft))
		require.False(t, status.IsValidTransition(domain.CampaignStatusPublished))
		require.False(t, status.IsValidTransition(domain.CampaignStatusActive))
		require.False(t, status.IsValidTransition(domain.CampaignStatusPaused))
		require.False(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})

	t.Run("completed cannot be resumed", func(t *testing.T) {
		status := domain.CampaignStatusCompleted
		require.False(t, status.IsValidTransition(domain.CampaignStatusActive),
			"completed should not resume to active")
	})
}

func TestCampaignLifecycle_StateMachine(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("full happy path lifecycle", func(t *testing.T) {
		status := domain.CampaignStatusDraft

		// Draft -> Published
		require.True(t, status.IsValidTransition(domain.CampaignStatusPublished))
		status = domain.CampaignStatusPublished

		// Published -> Active
		require.True(t, status.IsValidTransition(domain.CampaignStatusActive))
		status = domain.CampaignStatusActive

		// Active -> Completed
		require.True(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})

	t.Run("pause resume path", func(t *testing.T) {
		status := domain.CampaignStatusDraft
		status = domain.CampaignStatusPublished
		status = domain.CampaignStatusActive

		// Pause
		require.True(t, status.IsValidTransition(domain.CampaignStatusPaused))
		status = domain.CampaignStatusPaused

		// Resume
		require.True(t, status.IsValidTransition(domain.CampaignStatusActive))
		status = domain.CampaignStatusActive

		// Complete
		require.True(t, status.IsValidTransition(domain.CampaignStatusCompleted))
	})

	t.Run("cancellation from any non-terminal state", func(t *testing.T) {
		states := []domain.CampaignStatus{
			domain.CampaignStatusDraft,
			domain.CampaignStatusPublished,
			domain.CampaignStatusActive,
			domain.CampaignStatusPaused,
		}

		for _, s := range states {
			require.True(t, s.IsValidTransition(domain.CampaignStatusCancelled),
				"%s should be cancellable", s)
		}
	})
}