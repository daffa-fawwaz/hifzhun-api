package repositories

import (
	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookItemOverrideRepository interface {
	// FindByUserAndBookItemID returns the override for a specific user+bookItem pair.
	// Returns nil, nil when no override exists (not an error).
	FindByUserAndBookItemID(userID, bookItemID uuid.UUID) (*entities.BookItemOverride, error)

	// FindByUserAndBookItemIDs batch-fetches overrides for multiple book_item_ids.
	// Returns a map keyed by book_item_id string.
	FindByUserAndBookItemIDs(userID uuid.UUID, bookItemIDs []uuid.UUID) (map[uuid.UUID]*entities.BookItemOverride, error)

	// FindAllByUser returns every override that belongs to userID.
	FindAllByUser(userID uuid.UUID) ([]entities.BookItemOverride, error)

	// Upsert creates the override if it doesn't exist, updates it if it does.
	Upsert(override *entities.BookItemOverride) error

	// DeleteByUserAndBookItemID removes the personal override, restoring the
	// user's view to the canonical BookItem.
	DeleteByUserAndBookItemID(userID, bookItemID uuid.UUID) error
}

type bookItemOverrideRepository struct {
	db *gorm.DB
}

func NewBookItemOverrideRepository(db *gorm.DB) BookItemOverrideRepository {
	return &bookItemOverrideRepository{db: db}
}

func (r *bookItemOverrideRepository) FindByUserAndBookItemID(userID, bookItemID uuid.UUID) (*entities.BookItemOverride, error) {
	var override entities.BookItemOverride
	err := r.db.
		Where("user_id = ? AND book_item_id = ?", userID, bookItemID).
		First(&override).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &override, err
}

func (r *bookItemOverrideRepository) FindByUserAndBookItemIDs(userID uuid.UUID, bookItemIDs []uuid.UUID) (map[uuid.UUID]*entities.BookItemOverride, error) {
	result := make(map[uuid.UUID]*entities.BookItemOverride)
	if len(bookItemIDs) == 0 {
		return result, nil
	}

	var overrides []entities.BookItemOverride
	err := r.db.
		Where("user_id = ? AND book_item_id IN ?", userID, bookItemIDs).
		Find(&overrides).Error
	if err != nil {
		return nil, err
	}

	for i := range overrides {
		o := &overrides[i]
		result[o.BookItemID] = o
	}
	return result, nil
}

func (r *bookItemOverrideRepository) FindAllByUser(userID uuid.UUID) ([]entities.BookItemOverride, error) {
	var overrides []entities.BookItemOverride
	err := r.db.Where("user_id = ?", userID).Find(&overrides).Error
	return overrides, err
}

func (r *bookItemOverrideRepository) Upsert(override *entities.BookItemOverride) error {
	existing, err := r.FindByUserAndBookItemID(override.UserID, override.BookItemID)
	if err != nil {
		return err
	}
	if existing == nil {
		// Create new
		return r.db.Create(override).Error
	}
	// Update in-place, preserve ID
	override.ID = existing.ID
	override.CreatedAt = existing.CreatedAt
	return r.db.Save(override).Error
}

func (r *bookItemOverrideRepository) DeleteByUserAndBookItemID(userID, bookItemID uuid.UUID) error {
	return r.db.
		Where("user_id = ? AND book_item_id = ?", userID, bookItemID).
		Delete(&entities.BookItemOverride{}).Error
}
