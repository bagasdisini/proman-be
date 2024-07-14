package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
)

type UserHandler struct {
	userRepo *repository.UserCollRepository
	taskRepo *repository.TaskCollRepository
}

func NewUserHandler(e *echo.Echo, db *mongo.Database) *UserHandler {
	h := &UserHandler{
		userRepo: repository.NewUserRepository(db),
		taskRepo: repository.NewTaskRepository(db),
	}

	user := e.Group("/api", context.ContextHandler)

	user.GET("/user/count", h.userCount)

	return h
}

// User Count
// @Tags User
// @Summary Get user count
// @ID user-count
// @Router /api/user/count [get]
// @Accept json
// @Produce  json
// @Success 200
// @Security ApiKeyAuth
func (h *UserHandler) userCount(c echo.Context) error {
	count, err := h.taskRepo.CountActiveUsers(h.userRepo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, count)
}
