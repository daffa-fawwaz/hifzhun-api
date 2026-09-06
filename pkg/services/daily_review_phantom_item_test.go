package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

func TestDailyReviewPhantomItemFix(t *testing.T) {
	db := setupTestPostgresDB(t)

	itemRepo := repositories.NewItemRepository(db)
	reviewStateRepo := repositories.NewReviewStateRepository(db)
	dailyTaskRepo := repositories.NewDailyTaskRepository(db)
	juzRepo := repositories.NewJuzRepository(db)
	juzItemRepo := repositories.NewJuzItemRepository(db)

	dailyTaskSvc := services.NewDailyTaskService(
		reviewStateRepo,
		dailyTaskRepo,
		itemRepo,
		nil,
		nil,
		juzRepo,
		juzItemRepo,
	)

	userID := uuid.New()
	now := time.Now()
	taskDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Cleanup any previous test artifacts for this user
	defer func() {
		_ = db.Where("user_id = ?", userID).Delete(&entities.DailyTask{}).Error
		_ = db.Where("user_id = ?", userID).Delete(&entities.ReviewState{}).Error
		_ = db.Where("owner_id = ?", userID).Delete(&entities.Item{}).Error
	}()

	// -------------------------------------------------------------------------
	// Test C: Orphan Review State (review_state exists for non-existent item)
	// -------------------------------------------------------------------------
	nonExistentItemID := uuid.New()
	orphanReviewState := entities.ReviewState{
		ID:           uuid.New(),
		UserID:       userID,
		ItemID:       nonExistentItemID,
		State:        "active",
		Stability:    2.0,
		Difficulty:   5.0,
		NextReviewAt: &now,
	}
	if err := db.Create(&orphanReviewState).Error; err != nil {
		t.Fatalf("Failed to create orphan review state: %v", err)
	}

	// FindDueByUser MUST NOT return orphanReviewState because item does not exist in items table
	dueStates, err := reviewStateRepo.FindDueByUser(context.Background(), userID, now, 0)
	if err != nil {
		t.Fatalf("FindDueByUser error: %v", err)
	}
	for _, s := range dueStates {
		if s.ItemID == nonExistentItemID {
			t.Fatalf("Test C FAILED: FindDueByUser returned orphan review_state for item %s", nonExistentItemID)
		}
	}

	// GenerateToday MUST NOT create a daily_task for nonExistentItemID
	tasks, err := dailyTaskSvc.GenerateToday(context.Background(), userID, now, 0)
	if err != nil {
		t.Fatalf("GenerateToday error: %v", err)
	}
	for _, task := range tasks {
		if task.ItemID == nonExistentItemID {
			t.Fatalf("Test C FAILED: GenerateToday created daily_task for orphan item %s", nonExistentItemID)
		}
	}

	// -------------------------------------------------------------------------
	// Test A: Normal Item and Review State
	// -------------------------------------------------------------------------
	validItem := entities.Item{
		ID:         uuid.New(),
		OwnerID:    userID,
		SourceType: "quran",
		ContentRef: "surah:1:1-7",
		Status:     entities.ItemStatusFSRSActive,
	}
	if err := db.Create(&validItem).Error; err != nil {
		t.Fatalf("Failed to create valid item: %v", err)
	}

	validReviewState := entities.ReviewState{
		ID:           uuid.New(),
		UserID:       userID,
		ItemID:       validItem.ID,
		State:        "active",
		Stability:    2.0,
		Difficulty:   5.0,
		NextReviewAt: &now,
	}
	if err := db.Create(&validReviewState).Error; err != nil {
		t.Fatalf("Failed to create valid review state: %v", err)
	}

	dueStatesValid, err := reviewStateRepo.FindDueByUser(context.Background(), userID, now, 0)
	if err != nil {
		t.Fatalf("FindDueByUser error: %v", err)
	}
	foundValid := false
	for _, s := range dueStatesValid {
		if s.ItemID == validItem.ID {
			foundValid = true
			break
		}
	}
	if !foundValid {
		t.Fatalf("Test A FAILED: expected valid item to be returned by FindDueByUser")
	}

	// -------------------------------------------------------------------------
	// Test B: DeleteItemWithActiveState
	// -------------------------------------------------------------------------
	// Insert daily task for valid item
	validDailyTask := entities.DailyTask{
		ID:        uuid.New(),
		UserID:    userID,
		ItemID:    validItem.ID,
		CardID:    uuid.Nil,
		TaskDate:  taskDate,
		Source:    "quran",
		State:     "pending",
		CreatedAt: now,
	}
	if err := db.Create(&validDailyTask).Error; err != nil {
		t.Fatalf("Failed to create valid daily task: %v", err)
	}

	// Insert historical review log (MUST be preserved)
	historyLog := entities.ReviewLog{
		ID:         uuid.New(),
		UserID:     userID,
		ItemID:     validItem.ID,
		Rating:     3,
		ReviewedAt: now,
	}
	if err := db.Create(&historyLog).Error; err != nil {
		t.Fatalf("Failed to create review log: %v", err)
	}

	// Delete item with active state
	if err := itemRepo.DeleteItemWithActiveState(validItem.ID, userID); err != nil {
		t.Fatalf("DeleteItemWithActiveState error: %v", err)
	}

	// Verify item is gone
	var checkItem entities.Item
	if err := db.Where("id = ?", validItem.ID).First(&checkItem).Error; err == nil {
		t.Fatalf("Test B FAILED: item still exists after DeleteItemWithActiveState")
	}

	// Verify daily_task is gone
	var checkDailyTask entities.DailyTask
	if err := db.Where("user_id = ? AND item_id = ?", userID, validItem.ID).First(&checkDailyTask).Error; err == nil {
		t.Fatalf("Test B FAILED: daily_task still exists after DeleteItemWithActiveState")
	}

	// Verify review_state is gone
	var checkReviewState entities.ReviewState
	if err := db.Where("user_id = ? AND item_id = ?", userID, validItem.ID).First(&checkReviewState).Error; err == nil {
		t.Fatalf("Test B FAILED: review_state still exists after DeleteItemWithActiveState")
	}

	// Verify historical log is PRESERVED
	var checkHistoryLog entities.ReviewLog
	if err := db.Where("id = ?", historyLog.ID).First(&checkHistoryLog).Error; err != nil {
		t.Fatalf("Test B FAILED: historical review_log was unexpectedly deleted: %v", err)
	}

	t.Logf("✅ All Phantom Item Fix Tests PASSED successfully!")
}
