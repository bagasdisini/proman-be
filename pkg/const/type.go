package _const

// Project status
const (
	ProjectActive    = "active"
	ProjectCompleted = "completed"
	ProjectPending   = "pending"
	ProjectCancelled = "cancelled"
)

// Schedule type
const (
	ScheduleMeeting      = "meeting"
	ScheduleDiscussion   = "discussion"
	ScheduleReview       = "review"
	SchedulePresentation = "presentation"
	ScheduleEtc          = "etc"
)

// Task status
const (
	TaskActive    = "active"
	TaskTesting   = "testing"
	TaskCompleted = "completed"
	TaskCancelled = "cancelled"
)

var AllowedFileExtension = map[string]bool{
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"image/gif":       true,
	"image/bmp":       true,
	"image/webp":      true,
	"image/svg":       true,
	"txt/plain":       true,
	"application/pdf": true,
}
