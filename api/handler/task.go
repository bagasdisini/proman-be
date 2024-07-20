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

type TaskHandler struct {
	taskRepo *repository.TaskCollRepository
}

func NewTaskHandler(e *echo.Echo, db *mongo.Database) *TaskHandler {
	h := &TaskHandler{
		taskRepo: repository.NewTaskRepository(db),
	}

	task := e.Group("/api", context.ContextHandler)

	task.GET("/task/count", h.taskCount)

	return h
}

// Task Count
// @Tags Task
// @Summary Get task count
// @ID task-count
// @Router /api/task/count [get]
// @Accept json
// @Produce  json
// @Success 200
// @Security ApiKeyAuth
func (h *TaskHandler) taskCount(c echo.Context) error {
	currentEnd := time.Now()
	currentStart := currentEnd.AddDate(0, 0, -30)
	current, err := h.taskRepo.CountTask(currentStart, currentEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Task not found")
		}
		return err
	}

	prevEnd := currentStart.Add(-time.Second)
	prevStart := prevEnd.AddDate(0, 0, -30)
	previous, err := h.taskRepo.CountTask(prevStart, prevEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Task not found")
		}
		return err
	}

	docs := repository.CountTask{
		Current:   *current,
		LastMonth: *previous,
	}
	return c.JSON(http.StatusOK, docs)
}
