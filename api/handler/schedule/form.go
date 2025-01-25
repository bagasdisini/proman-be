package schedule

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"proman-backend/internal/pkg/log"
	"strings"
	"time"
)

const (
	minNameLength        = 1
	maxNameLength        = 100
	minDescriptionLength = 10
	maxDescriptionLength = 1000
	minTypeLength        = 1
	maxTypeLength        = 50
)

type scheduleForm struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
	Contributor string `json:"contributor" form:"contributor"`
	Type        string `json:"type" form:"type"`
}

type errorDoc struct {
	Field   string `json:"field"`
	Message string `json:"message"`
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

	var validationErrors []errorDoc

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
	if len(form.Type) < minTypeLength || len(form.Type) > maxTypeLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "type",
			Message: "Type must be between 1 and 50 characters.",
		})
	}

	// Validate start date
	if form.StartDate <= 0 {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "start_date",
			Message: "Invalid start date.",
		})
	} else {
		startDate := time.Unix(form.StartDate, 0)
		if startDate.After(time.Now()) {
			validationErrors = append(validationErrors, errorDoc{
				Field:   "start_date",
				Message: "Start date cannot be in the future.",
			})
		}
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

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}
