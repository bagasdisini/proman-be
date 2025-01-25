package schedule

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/log"
	"proman-backend/internal/pkg/util"
	"strings"
	"time"
)

type Handler struct {
	scheduleRepo *repository.ScheduleCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		scheduleRepo: repository.NewScheduleRepository(db),
	}

	schedule := e.Group("/api", context.ContextHandler)

	schedule.GET("/schedules", h.list)

	schedule.POST("/schedule", h.create)

	return h
}

// List Schedule
// @Tags Schedule
// @Summary Get list of schedule
// @ID list-schedule
// @Router /api/schedules [get]
// @Param q query string false "Search by name"
// @Param type query string false "Search by type" Enums(all, meeting, discussion, review, presentation, etc)
// @Param start query string false "Start date"
// @Param end query string false "End date"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) list(c echo.Context) error {
	cq := util.NewCommonQuery(c)

	schedules, err := h.scheduleRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Schedule not found")
		}
		log.Errorf("Error finding schedule: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, schedules)
}

// Create Schedule
// @Tags Schedule
// @Summary Create schedule
// @ID schedule-create
// @Router /api/schedule [post]
// @Accept json
// @Produce json
// @Param body body scheduleForm true "Schedule data"
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) create(c echo.Context) error {
	form, err := newScheduleForm(c)
	if err != nil {
		return err
	}

	contributorsOId := make([]bson.ObjectID, 0)
	for _, user := range strings.Split(form.Contributor, ",") {
		userOId, err := bson.ObjectIDFromHex(user)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor.")
		}
		contributorsOId = append(contributorsOId, userOId)
	}

	schedule := &repository.Schedule{
		ID:          bson.NewObjectID(),
		Name:        form.Name,
		Description: form.Description,
		StartDate:   time.UnixMilli(form.StartDate),
		EndDate:     time.UnixMilli(form.EndDate),
		Contributor: contributorsOId,
		Type:        form.Type,
		CreatedAt:   time.Now(),
		IsDeleted:   false,
	}

	if err := h.scheduleRepo.CreateOne(schedule); err != nil {
		log.Errorf("Failed to create schedule: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusCreated, schedule)
}
