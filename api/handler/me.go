package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
	"proman-backend/pkg/util"
)

type MeHandler struct {
	userRepo     *repository.UserCollRepository
	projectRepo  *repository.ProjectCollRepository
	taskRepo     *repository.TaskCollRepository
	scheduleRepo *repository.ScheduleCollRepository
}

func NewMeHandler(e *echo.Echo, db *mongo.Database) *MeHandler {
	h := &MeHandler{
		userRepo:     repository.NewUserRepository(db),
		projectRepo:  repository.NewProjectRepository(db),
		taskRepo:     repository.NewTaskRepository(db),
		scheduleRepo: repository.NewScheduleRepository(db),
	}

	me := e.Group("/api", context.ContextHandler)

	me.GET("/me", h.me)
	me.GET("/me/schedule", h.mySchedule)

	return h
}

// Me
// @Tags Me
// @Summary Get my info
// @ID me
// @Router /api/me [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *MeHandler) me(c echo.Context) error {
	uc := c.(*context.Context)

	user, err := h.userRepo.FindOneByID(uc.Claims.IDAsObjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}
		return err
	}

	if user.IsDeleted {
		return echo.NewHTTPError(http.StatusUnauthorized, "User is deleted")
	}

	return c.JSON(http.StatusOK, user)
}

// My Schedule
// @Tags Me
// @Summary Get my schedule
// @ID my-schedule
// @Router /api/me/schedule [get]
// @Param q query string false "Search by name"
// @Param type query string false "Search by type" Enums(all, meeting, discussion, review, presentation, etc)
// @Param start query string false "Start date"
// @Param end query string false "End date"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *MeHandler) mySchedule(c echo.Context) error {
	cq, err := util.NewScheduleQuery(c)
	if err != nil {
		return err
	}

	uc := c.(*context.Context)
	cq.Contributor = []primitive.ObjectID{uc.Claims.IDAsObjectID}

	schedules, err := h.scheduleRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Schedule not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, schedules)
}
