package _const

// Project status
const (
	ProjectActive    = "active"
	ProjectCompleted = "completed"
	ProjectPending   = "pending"
	ProjectCancelled = "cancelled"
)

func IsValidProjectStatus(projectStatus string) bool {
	switch projectStatus {
	case ProjectActive, ProjectCompleted, ProjectPending, ProjectCancelled:
		return true
	}
	return false
}

// Project type
const (
	ProjectFrontend = "frontend"
	ProjectBackend  = "backend"
	ProjectMobile   = "mobile"
	ProjectDesktop  = "desktop"
	ProjectMonitor  = "monitor"
	ProjectTool     = "tool"
	ProjectEtc      = "etc"
)

func IsValidProjectType(projectType string) bool {
	switch projectType {
	case ProjectFrontend, ProjectBackend, ProjectMobile, ProjectDesktop, ProjectMonitor, ProjectTool, ProjectEtc:
		return true
	}
	return false
}

// Schedule type
const (
	ScheduleMeeting      = "meeting"
	ScheduleDiscussion   = "discussion"
	ScheduleReview       = "review"
	SchedulePresentation = "presentation"
	ScheduleEtc          = "etc"
)

func IsValidScheduleType(scheduleType string) bool {
	switch scheduleType {
	case ScheduleMeeting, ScheduleDiscussion, ScheduleReview, SchedulePresentation, ScheduleEtc:
		return true
	}
	return false
}

// Task status
const (
	TaskActive    = "active"
	TaskTesting   = "testing"
	TaskCompleted = "completed"
	TaskCancelled = "cancelled"
)

func IsValidTaskStatus(taskStatus string) bool {
	switch taskStatus {
	case TaskActive, TaskTesting, TaskCompleted, TaskCancelled:
		return true
	}
	return false
}

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

func IsValidFileExtension(fileType string) bool {
	return AllowedFileExtension[fileType]
}
