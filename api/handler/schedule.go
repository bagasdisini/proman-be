package handler

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
	"strings"
	"time"
)

type scheduleForm struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
	Contributor string `json:"contributor" form:"contributor"`
	Type        string `json:"type" form:"type"`
}

func newScheduleForm(c echo.Context) (*scheduleForm, error) {
	form := scheduleForm{}
	if err := c.Bind(&form); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid data format.")
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Description = strings.TrimSpace(form.Description)
	form.Contributor = strings.TrimSpace(form.Contributor)
	form.Type = strings.TrimSpace(form.Type)

	if form.Name == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Name cannot be empty.")
	}
	if form.Description == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Description cannot be empty.")
	}
	if form.Contributor == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Contributor cannot be empty.")
	}
	if form.Type == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Type cannot be empty.")
	}
	if form.StartDate <= 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid start date.")
	}
	if form.EndDate <= 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid end date.")
	}
	return &form, nil
}

type ScheduleHandler struct {
	scheduleRepo *repository.ScheduleCollRepository
}

func NewScheduleHandler(e *echo.Echo, db *mongo.Database) *ScheduleHandler {
	h := &ScheduleHandler{
		scheduleRepo: repository.NewScheduleRepository(db),
	}

	schedule := e.Group("/api", context.ContextHandler)

	schedule.POST("/schedule", h.create)

	return h
}

// Create Schedule
// @Tags Schedule
// @Summary Create schedule
// @ID schedule-create
// @Router /api/schedule [post]
// @Accept json
// @Produce  json
// @Param body body scheduleForm true "Schedule data"
// @Success 200
// @Security ApiKeyAuth
func (h *ScheduleHandler) create(c echo.Context) error {
	form, err := newScheduleForm(c)
	if err != nil {
		return err
	}

	contributorsOId := make([]primitive.ObjectID, 0)
	for _, user := range strings.Split(form.Contributor, ",") {
		userOId, err := primitive.ObjectIDFromHex(user)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor.")
		}
		contributorsOId = append(contributorsOId, userOId)
	}

	schedule := &repository.Schedule{
		ID:          primitive.NewObjectID(),
		Name:        form.Name,
		Description: form.Description,
		StartDate:   time.Unix(form.StartDate, 0),
		EndDate:     time.Unix(form.EndDate, 0),
		Contributor: contributorsOId,
		Type:        form.Type,
		CreatedAt:   time.Now(),
		IsDeleted:   false,
	}

	if err := h.scheduleRepo.CreateOne(schedule); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, schedule)
}
