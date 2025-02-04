package schedule

import (
	"github.com/labstack/echo/v4"
	"net/http"
	_const "proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/log"
	"strings"
	"time"
)

const (
	minNameLength        = 1
	maxNameLength        = 100
	minDescriptionLength = 10
	maxDescriptionLength = 1000
)

type errorDoc struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type scheduleForm struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
	StartTime   string `json:"start_time" form:"start_time"`
	EndTime     string `json:"end_time" form:"end_time"`
	Contributor string `json:"contributor" form:"contributor"`
	Type        string `json:"type" form:"type"`
}

func newScheduleForm(c echo.Context) (*scheduleForm, error) {
	form := new(scheduleForm)
	if err := c.Bind(form); err != nil {
		log.Errorf("Error binding schedule form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid data format.")
	}

	// Sanitize inputs
	form.Name = strings.TrimSpace(form.Name)
	form.Description = strings.TrimSpace(form.Description)
	form.Contributor = strings.TrimSpace(form.Contributor)
	form.Type = strings.TrimSpace(form.Type)
	form.StartTime = strings.TrimSpace(form.StartTime)
	form.EndTime = strings.TrimSpace(form.EndTime)

	validationErrors := make([]errorDoc, 0)

	// Validate name
	if len(form.Name) < minNameLength || len(form.Name) > maxNameLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "name",
			Message: "Name must be between 1 and 100 characters.",
		})
	}

	// Validate description
	if len(form.Description) < minDescriptionLength || len(form.Description) > maxDescriptionLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "description",
			Message: "Description must be between 10 and 1000 characters.",
		})
	}

	// Validate contributor
	if form.Contributor == "" {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "contributor",
			Message: "Contributor cannot be empty.",
		})
	}

	// Validate type
	if !_const.IsValidScheduleType(form.Type) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "type",
			Message: "Invalid schedule type.",
		})
	}

	// Validate start date
	if form.StartDate <= 0 {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "start_date",
			Message: "Invalid start date.",
		})
	}

	// Validate end date
	if form.EndDate <= 0 {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "end_date",
			Message: "Invalid end date.",
		})
	} else if form.EndDate <= form.StartDate {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "end_date",
			Message: "End date must be after the start date.",
		})
	}

	// Validate start time
	if form.StartTime == "" {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "start_time",
			Message: "Start time cannot be empty.",
		})
	} else if _, err := time.Parse("15:04", form.StartTime); err != nil {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "start_time",
			Message: "Invalid start time format.",
		})
	}

	// Validate end time
	if form.EndTime == "" {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "end_time",
			Message: "End time cannot be empty.",
		})
	} else if _, err := time.Parse("15:04", form.EndTime); err != nil {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "end_time",
			Message: "Invalid end time format.",
		})
	} else if form.EndTime <= form.StartTime {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "end_time",
			Message: "End time must be after the start time.",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}
