package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
	"time"
)

type UserHandler struct {
	userRepo    *repository.UserCollRepository
	taskRepo    *repository.TaskCollRepository
	projectRepo *repository.ProjectCollRepository
}

func NewUserHandler(e *echo.Echo, db *mongo.Database) *UserHandler {
	h := &UserHandler{
		userRepo:    repository.NewUserRepository(db),
		taskRepo:    repository.NewTaskRepository(db),
		projectRepo: repository.NewProjectRepository(db),
	}

	user := e.Group("/api", context.ContextHandler)

	user.GET("/user", h.userList)
	user.GET("/user/count", h.userCount)

	return h
}

// User Count
// @Tags User
// @Summary Get user count
// @ID user-count
// @Router /api/user/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *UserHandler) userCount(c echo.Context) error {
	currentEnd := time.Date(time.Now().Year(), 12, 31, 23, 59, 59, 0, time.Local)
	currentStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Local)
	current, err := h.taskRepo.CountUserThatHaveTask(h.userRepo, currentStart, currentEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		return err
	}

	prevEnd := currentStart.AddDate(0, 0, -1)
	prevStart := time.Date(prevEnd.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	previous, err := h.taskRepo.CountUserThatHaveTask(h.userRepo, prevStart, prevEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		return err
	}

	docs := repository.CountUser{
		Current:  *current,
		LastYear: *previous,
	}
	return c.JSON(http.StatusOK, docs)
}

// User List
// @Tags User
// @Summary Get list users
// @ID user-latest
// @Router /api/user [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *UserHandler) userList(c echo.Context) error {
	users, err := h.userRepo.FindAllUsers(h.projectRepo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, users)
}
