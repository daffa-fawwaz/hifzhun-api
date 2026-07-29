package services

import (
	"time"

	"hifzhun-api/pkg/entities"
)

// MinimumReviewsToGraduate is required regardless of whether graduation is
// evaluated immediately after a review or while generating daily tasks.
const MinimumReviewsToGraduate = 5

func qualifiesForGraduation(item *entities.Item, now time.Time) bool {
	if item == nil || item.Status != entities.ItemStatusFSRSActive || item.ReviewCount < MinimumReviewsToGraduate {
		return false
	}

	daysInFSRSActive := 0
	if item.FSRSStartAt != nil {
		daysInFSRSActive = int(now.Sub(*item.FSRSStartAt).Hours() / 24)
	} else if item.IntervalEndAt != nil {
		daysInFSRSActive = int(now.Sub(*item.IntervalEndAt).Hours() / 24)
	}

	return daysInFSRSActive >= entities.GraduationIntervalDays ||
		item.Stability >= entities.GraduateStabilityThreshold
}
