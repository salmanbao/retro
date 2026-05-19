package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

func TestCreativeBriefValidation(t *testing.T) {
	t.Run("valid brief passes validation", func(t *testing.T) {
		brief := &domain.CreativeBrief{
			CampaignID:             uuid.New(),
			KeyMessages:            []string{"message 1", "message 2"},
			RequiredHashtags:       []string{"#brand", "#ad"},
			MandatoryTalkingPoints: []string{"point 1"},
		}
		err := brief.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing campaign ID fails validation", func(t *testing.T) {
		brief := &domain.CreativeBrief{
			CampaignID:             uuid.Nil,
			KeyMessages:            []string{"message 1"},
			RequiredHashtags:       []string{"#brand"},
			MandatoryTalkingPoints: []string{"point 1"},
		}
		err := brief.Validate()
		assert.ErrorIs(t, err, domain.ErrBriefCampaignRequired)
	})

	t.Run("empty key messages fails validation", func(t *testing.T) {
		brief := &domain.CreativeBrief{
			CampaignID:             uuid.New(),
			KeyMessages:            []string{},
			RequiredHashtags:       []string{"#brand"},
			MandatoryTalkingPoints: []string{"point 1"},
		}
		err := brief.Validate()
		assert.ErrorIs(t, err, domain.ErrBriefKeyMessagesEmpty)
	})

	t.Run("empty required hashtags fails validation", func(t *testing.T) {
		brief := &domain.CreativeBrief{
			CampaignID:             uuid.New(),
			KeyMessages:            []string{"message 1"},
			RequiredHashtags:       []string{},
			MandatoryTalkingPoints: []string{"point 1"},
		}
		err := brief.Validate()
		assert.ErrorIs(t, err, domain.ErrBriefHashtagsRequired)
	})

	t.Run("empty mandatory talking points fails validation", func(t *testing.T) {
		brief := &domain.CreativeBrief{
			CampaignID:             uuid.New(),
			KeyMessages:            []string{"message 1"},
			RequiredHashtags:       []string{"#brand"},
			MandatoryTalkingPoints: []string{},
		}
		err := brief.Validate()
		assert.ErrorIs(t, err, domain.ErrBriefTalkingPointsReq)
	})
}

func TestCreativeBriefEditRestrictions(t *testing.T) {
	t.Run("draft campaign allows full edit", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		assert.True(t, brief.CanEditFull(domain.CampaignStatusDraft))
	})

	t.Run("paused campaign allows full edit", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		assert.True(t, brief.CanEditFull(domain.CampaignStatusPaused))
	})

	t.Run("published campaign restricts edit", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		assert.False(t, brief.CanEditFull(domain.CampaignStatusPublished))
	})

	t.Run("active campaign restricts edit", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		assert.False(t, brief.CanEditFull(domain.CampaignStatusActive))
	})

	t.Run("completed campaign restricts edit", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		assert.False(t, brief.CanEditFull(domain.CampaignStatusCompleted))
	})

	t.Run("cancelled campaign restricts edit", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		assert.False(t, brief.CanEditFull(domain.CampaignStatusCancelled))
	})

	t.Run("restricted edit fields are correct", func(t *testing.T) {
		brief := &domain.CreativeBrief{}
		fields := brief.RestrictedEditFields()
		assert.Equal(t, []string{
			"tone_and_style_guidelines",
			"target_audience_description",
			"example_video_links",
		}, fields)
	})
}
