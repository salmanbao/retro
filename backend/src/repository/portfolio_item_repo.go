package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// PortfolioItemDB implements PortfolioItemRepository using GORM.
type PortfolioItemDB struct {
	db *gorm.DB
}

// NewPortfolioItemDB creates a new PortfolioItemDB.
func NewPortfolioItemDB(db *gorm.DB) *PortfolioItemDB {
	return &PortfolioItemDB{db: db}
}

// Create inserts a new portfolio item.
func (r *PortfolioItemDB) Create(ctx context.Context, item *domain.PortfolioItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// ByID retrieves a portfolio item by ID (excludes soft-deleted).
func (r *PortfolioItemDB) ByID(ctx context.Context, id uuid.UUID) (*domain.PortfolioItem, error) {
	var item domain.PortfolioItem
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPortfolioItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

// ByProfileID retrieves all portfolio items for a profile ordered by display_order, then created_at.
func (r *PortfolioItemDB) ByProfileID(ctx context.Context, profileID uuid.UUID, limit, offset int) ([]*domain.PortfolioItem, error) {
	var items []*domain.PortfolioItem
	err := r.db.WithContext(ctx).
		Where("profile_id = ? AND deleted_at IS NULL", profileID).
		Order("display_order ASC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Update updates an existing portfolio item.
func (r *PortfolioItemDB) Update(ctx context.Context, item *domain.PortfolioItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete soft-deletes a portfolio item.
func (r *PortfolioItemDB) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&domain.PortfolioItem{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPortfolioItemNotFound
	}
	return nil
}

// CountByProfileID counts active portfolio items for a profile.
func (r *PortfolioItemDB) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.PortfolioItem{}).Where("profile_id = ? AND deleted_at IS NULL", profileID).Count(&count).Error
	return count, err
}
