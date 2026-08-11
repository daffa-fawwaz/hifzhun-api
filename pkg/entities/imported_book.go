package entities

import (
	"time"

	"github.com/google/uuid"
)

type ImportedBook struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_book,unique" json:"user_id"`
	BookID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_book,unique" json:"book_id"`
	CreatedAt time.Time `json:"created_at"`
}
