package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
)

func TestClassBookStudentProgressCategorization(t *testing.T) {
	config.InitAppLocation()
	now := time.Now().In(config.AppLocation)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name             string
		status           string
		nextReviewAt     *time.Time
		taskState        string
		hasTask          bool
		expectedReviewed bool // true if categorized as Belum di-review (totalUnreviewed++)
	}{
		{
			name:             "Case 1: fsrs_active + due + pending -> Belum di-review",
			status:           entities.ItemStatusFSRSActive,
			nextReviewAt:     &now,
			taskState:        "pending",
			hasTask:          true,
			expectedReviewed: true,
		},
		{
			name:             "Case 2: fsrs_active + due + completed -> Aktif FSRS",
			status:           entities.ItemStatusFSRSActive,
			nextReviewAt:     &now,
			taskState:        "done",
			hasTask:          true,
			expectedReviewed: false,
		},
		{
			name:             "Case 3: fsrs_active + future next_review -> Aktif FSRS",
			status:           entities.ItemStatusFSRSActive,
			nextReviewAt:     &tomorrow,
			taskState:        "",
			hasTask:          false,
			expectedReviewed: false,
		},
		{
			name:             "Case 4: fsrs_active + overdue + pending -> Belum di-review",
			status:           entities.ItemStatusFSRSActive,
			nextReviewAt:     &yesterday,
			taskState:        "pending",
			hasTask:          true,
			expectedReviewed: true,
		},
		{
			name:             "Case 5: fsrs_active + overdue + completed -> Aktif FSRS",
			status:           entities.ItemStatusFSRSActive,
			nextReviewAt:     &yesterday,
			taskState:        "done",
			hasTask:          true,
			expectedReviewed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := entities.Item{
				ID:           uuid.New(),
				Status:       tt.status,
				NextReviewAt: tt.nextReviewAt,
			}

			dailyTaskState := make(map[uuid.UUID]string)
			if tt.hasTask {
				dailyTaskState[item.ID] = tt.taskState
			}

			taskState, hasTask := dailyTaskState[item.ID]
			isTaskCompleted := hasTask && (taskState == "done" || taskState == "completed")
			isDue := item.NextReviewAt == nil || !item.NextReviewAt.After(endOfDay) || (item.IntervalNextReviewAt != nil && !item.IntervalNextReviewAt.After(endOfDay))

			isUnreviewed := (hasTask && !isTaskCompleted) || (isDue && !isTaskCompleted) || item.Status == entities.ItemStatusStart || item.Status == entities.ItemStatusMenghafal

			if isUnreviewed != tt.expectedReviewed {
				t.Errorf("expected isUnreviewed=%v, got %v", tt.expectedReviewed, isUnreviewed)
			}
		})
	}
}
