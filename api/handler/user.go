package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
	"proman-backend/pkg/log"
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
	count, err := h.taskRepo.CountUserTask(h.userRepo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		log.Errorf("Error counting user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, count)
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
	users, err := h.userRepo.FindAllUsers()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, users)
}
