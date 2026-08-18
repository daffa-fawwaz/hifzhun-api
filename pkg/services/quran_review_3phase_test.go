package services_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

func setupTestPostgresDB(t *testing.T) *gorm.DB {
	config.InitAppLocation()
	_ = godotenv.Load("../../.env")

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "daffafawwaz"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "root"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "hifzhun_db"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping test: database connection failed: %v", err)
	}

	return db
}

// Test A & B & C: Legacy Quran Personal Item with status = 'interval'
func TestLegacyQuranPersonalIntervalReviewFlow(t *testing.T) {
	db := setupTestPostgresDB(t)
	now := time.Now().In(config.AppLocation)
	yesterday := now.AddDate(0, 0, -1)

	userID := uuid.New()
	juzID := uuid.New()

	// 1. Create personal Juz (class_id IS NULL)
	juz := &entities.Juz{
		ID:       juzID,
		UserID:   userID,
		ClassID:  nil,
		Index:    29,
		IsActive: true,
	}
	if err := db.Create(juz).Error; err != nil {
		t.Fatalf("failed to create test juz: %v", err)
	}
	defer db.Delete(juz)

	// 2. Create legacy Quran Personal item (status = 'interval')
	intervalNext := yesterday
	item := &entities.Item{
		ID:                   uuid.New(),
		OwnerID:              userID,
		SourceType:           "quran",
		ContentRef:           "surah:78:1-5",
		Status:               entities.ItemStatusInterval,
		IntervalNextReviewAt: &intervalNext,
		Stability:            0,
		Difficulty:           5.0,
		CreatedAt:            now.AddDate(0, 0, -30),
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("failed to create test item: %v", err)
	}
	defer db.Delete(item)

	juzItem := &entities.JuzItem{
		ID:     uuid.New(),
		JuzID:  juzID,
		ItemID: item.ID,
	}
	if err := db.Create(juzItem).Error; err != nil {
		t.Fatalf("failed to create test juz_item: %v", err)
	}
	defer db.Delete(juzItem)

	// Test A: DailyTaskService.GenerateToday finds the legacy item and creates daily task
	itemRepo := repositories.NewItemRepository(db)
	juzRepo := repositories.NewJuzRepository(db)
	juzItemRepo := repositories.NewJuzItemRepository(db)
	dailyTaskRepo := repositories.NewDailyTaskRepository(db)
	reviewStateRepo := repositories.NewReviewStateRepository(db)

	dailyService := services.NewDailyTaskService(
		reviewStateRepo,
		dailyTaskRepo,
		itemRepo,
		nil,
		nil,
		juzRepo,
		juzItemRepo,
	)

	tasks, err := dailyService.GenerateToday(context.Background(), userID, now, 0)
	if err != nil {
		t.Fatalf("GenerateToday error: %v", err)
	}

	foundTask := false
	for _, task := range tasks {
		if task.ItemID == item.ID {
			foundTask = true
			if task.Source != "interval_review" {
				t.Errorf("expected task.Source 'interval_review', got '%s'", task.Source)
			}
		}
	}
	if !foundTask {
		t.Errorf("Test A FAILED: expected GenerateToday to find legacy interval item")
	}

	// Test B & C: ReviewItem normalizes legacy interval item to fsrs_active and calculates FSRS review
	reviewService := services.NewItemReviewService(
		itemRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		juzItemRepo,
	)

	reviewResult, err := reviewService.ReviewItem(userID, item.ID, fsrs.Good, now)
	if err != nil {
		t.Fatalf("Test B FAILED: ReviewItem returned error for legacy interval item: %v", err)
	}

	if reviewResult.Item.Status != entities.ItemStatusFSRSActive {
		t.Errorf("Test C FAILED: expected item status 'fsrs_active', got '%s'", reviewResult.Item.Status)
	}

	if reviewResult.Item.NextReviewAt == nil {
		t.Errorf("Test C FAILED: expected NextReviewAt to be non-nil after review")
	}

	if reviewResult.Item.Stability <= 0 {
		t.Errorf("Test C FAILED: expected FSRS stability > 0, got %f", reviewResult.Item.Stability)
	}
}

// Test D: New Quran Personal Item (menghafal -> fsrs_active)
func TestNewQuranPersonalActivationAndReview(t *testing.T) {
	db := setupTestPostgresDB(t)
	now := time.Now().In(config.AppLocation)

	userID := uuid.New()
	itemRepo := repositories.NewItemRepository(db)
	juzItemRepo := repositories.NewJuzItemRepository(db)

	statusService := services.NewItemStatusService(itemRepo, nil, nil, nil)

	item := &entities.Item{
		ID:         uuid.New(),
		OwnerID:    userID,
		SourceType: "quran",
		ContentRef: "surah:1:1-7",
		Status:     entities.ItemStatusMenghafal,
		CreatedAt:  now,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	defer db.Delete(item)

	// Activate to FSRS
	activatedItem, err := statusService.ActivateToFSRS(item.ID, userID)
	if err != nil {
		t.Fatalf("ActivateToFSRS error: %v", err)
	}

	if activatedItem.Status != entities.ItemStatusFSRSActive {
		t.Errorf("expected status 'fsrs_active', got '%s'", activatedItem.Status)
	}

	if activatedItem.NextReviewAt == nil {
		t.Errorf("expected NextReviewAt to be set")
	}

	// Review item
	reviewService := services.NewItemReviewService(itemRepo, nil, nil, nil, nil, nil, juzItemRepo)
	res, err := reviewService.ReviewItem(userID, item.ID, fsrs.Good, now.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("ReviewItem error: %v", err)
	}

	if res.Item.ReviewCount != 1 {
		t.Errorf("expected ReviewCount 1, got %d", res.Item.ReviewCount)
	}
}
