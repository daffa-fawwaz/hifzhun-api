package repositories

import (
	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookItemRepository interface {
	Create(item *entities.BookItem) error
	FindByID(id string) (*entities.BookItem, error)
	FindByIDs(ids []string) ([]entities.BookItem, error)
	// FindByBookID returns only canonical items (importer_id IS NULL).
	FindByBookID(bookID string) ([]entities.BookItem, error)
	// FindByBookIDForImporter returns canonical items + items created by importerID.
	FindByBookIDForImporter(bookID string, importerID uuid.UUID) ([]entities.BookItem, error)
	// FindByModuleID returns only canonical items (importer_id IS NULL).
	FindByModuleID(moduleID string) ([]entities.BookItem, error)
	// FindByModuleIDForImporter returns canonical items + items created by importerID.
	FindByModuleIDForImporter(moduleID string, importerID uuid.UUID) ([]entities.BookItem, error)
	// FindImporterItems returns only items created by the given importer for a book.
	FindImporterItems(bookID string, importerID uuid.UUID) ([]entities.BookItem, error)
	Update(item *entities.BookItem) error
	Delete(id string) error
	DeleteByBookID(bookID string) error
	DeleteByModuleID(moduleID string) error
}

type bookItemRepository struct {
	db *gorm.DB
}

func NewBookItemRepository(db *gorm.DB) BookItemRepository {
	return &bookItemRepository{db}
}

func (r *bookItemRepository) Create(item *entities.BookItem) error {
	return r.db.Create(item).Error
}

func (r *bookItemRepository) FindByID(id string) (*entities.BookItem, error) {
	var item entities.BookItem
	err := r.db.Where("id = ?", id).First(&item).Error
	return &item, err
}

func (r *bookItemRepository) FindByIDs(ids []string) ([]entities.BookItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []entities.BookItem
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

// FindByBookID returns only canonical items (importer_id IS NULL).
// This is what the book owner and all other users see.
func (r *bookItemRepository) FindByBookID(bookID string) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("book_id = ? AND importer_id IS NULL", bookID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

// FindByBookIDForImporter returns canonical items PLUS items created by importerID.
// Used when an importer views the book they imported — they see canonical items and
// their own personal additions, but not items added by other importers.
func (r *bookItemRepository) FindByBookIDForImporter(bookID string, importerID uuid.UUID) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("book_id = ? AND (importer_id IS NULL OR importer_id = ?)", bookID, importerID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

// FindByModuleID returns only canonical items (importer_id IS NULL).
func (r *bookItemRepository) FindByModuleID(moduleID string) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("module_id = ? AND importer_id IS NULL", moduleID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

// FindByModuleIDForImporter returns canonical items + items created by importerID.
func (r *bookItemRepository) FindByModuleIDForImporter(moduleID string, importerID uuid.UUID) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("module_id = ? AND (importer_id IS NULL OR importer_id = ?)", moduleID, importerID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

// FindImporterItems returns only items created by importerID for a specific book.
func (r *bookItemRepository) FindImporterItems(bookID string, importerID uuid.UUID) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("book_id = ? AND importer_id = ?", bookID, importerID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

func (r *bookItemRepository) Update(item *entities.BookItem) error {
	return r.db.Save(item).Error
}

func (r *bookItemRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.BookItem{}).Error
}

func (r *bookItemRepository) DeleteByBookID(bookID string) error {
	return r.db.Where("book_id = ?", bookID).Delete(&entities.BookItem{}).Error
}

func (r *bookItemRepository) DeleteByModuleID(moduleID string) error {
	return r.db.Where("module_id = ?", moduleID).Delete(&entities.BookItem{}).Error
}
