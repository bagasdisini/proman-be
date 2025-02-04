package task

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
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
)

type errorDoc struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type taskForm struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
	Contributor string `json:"contributor" form:"contributor"`
	ProjectID   string `json:"project_id" form:"project_id"`
}

func newTaskForm(c echo.Context) (*taskForm, error) {
	form := new(taskForm)
	if err := c.Bind(form); err != nil {
		log.Errorf("Error binding task form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid data format.")
	}

	// Sanitize inputs
	form.Name = strings.TrimSpace(form.Name)
	form.Description = strings.TrimSpace(form.Description)
	form.Contributor = strings.TrimSpace(form.Contributor)
	form.ProjectID = strings.TrimSpace(form.ProjectID)

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

	// Validate project ID
	if form.ProjectID == "" {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "project_id",
			Message: "Project ID cannot be empty.",
		})
	} else if _, err := bson.ObjectIDFromHex(form.ProjectID); err != nil {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "project_id",
			Message: "Invalid project ID.",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}
