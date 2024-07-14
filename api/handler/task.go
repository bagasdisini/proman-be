package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
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
	count, err := h.taskRepo.CountTask()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Task not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, count)
}
