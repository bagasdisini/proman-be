package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
	"proman-backend/pkg/util"
	"time"
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
	currentEnd := time.Now()
	currentStart := currentEnd.AddDate(0, 0, -30)
	current, err := h.taskRepo.CountUserThatHaveTask(h.userRepo, currentStart, currentEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		return err
	}

	prevEnd := currentStart.Add(-time.Second)
	prevStart := prevEnd.AddDate(0, 0, -30)
	previous, err := h.taskRepo.CountUserThatHaveTask(h.userRepo, prevStart, prevEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		return err
	}

	docs := repository.CountUser{
		Current:   *current,
		LastMonth: *previous,
	}
	return c.JSON(http.StatusOK, docs)
}

// User Latest
// @Tags User
// @Summary Get latest users joined
// @ID user-latest
// @Param limit query int false "Limit" default(10)
// @Param sort query string false "Sort" default("desc") Enums(desc,asc)
// @Router /api/user/latest [get]
// @Accept json
// @Produce json
//@Success 200
// @Security ApiKeyAuth
func (h *UserHandler) userList(c echo.Context) error {
	_, err := util.NewCommonQuery(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return err
}
