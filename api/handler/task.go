package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	_const "proman-backend/pkg/const"
	"proman-backend/pkg/context"
	"strings"
	"time"
)

type taskForm struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
	Contributor string `json:"contributor" form:"contributor"`
	ProjectID   string `json:"project_id" form:"project_id"`
}

func newTaskForm(c echo.Context) (*taskForm, error) {
	form := taskForm{}
	if err := c.Bind(&form); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid data format.")
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Description = strings.TrimSpace(form.Description)
	form.Contributor = strings.TrimSpace(form.Contributor)
	form.ProjectID = strings.TrimSpace(form.ProjectID)

	if form.Name == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Name cannot be empty.")
	}
	if form.Description == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Description cannot be empty.")
	}
	if form.Contributor == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Contributor cannot be empty.")
	}
	if form.StartDate <= 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid start date.")
	}
	if form.EndDate <= 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid end date.")
	}
	_, err := primitive.ObjectIDFromHex(form.ProjectID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}
	return &form, nil
}

type TaskHandler struct {
	taskRepo *repository.TaskCollRepository
}

func NewTaskHandler(e *echo.Echo, db *mongo.Database) *TaskHandler {
	h := &TaskHandler{
		taskRepo: repository.NewTaskRepository(db),
	}

	task := e.Group("/api", context.ContextHandler)

	task.GET("/task/count", h.taskCount)

	task.POST("/task", h.create)

	task.DELETE("/task/:id", h.delete)

	return h
}

// Task Count
// @Tags Task
// @Summary Get task count
// @ID task-count
// @Router /api/task/count [get]
// @Accept json
// @Produce json
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

// Create Task
// @Tags Task
// @Summary Create task
// @ID task-create
// @Router /api/task [post]
// @Accept json
// @Produce json
// @Param body body taskForm true "Task data"
// @Success 200
// @Security ApiKeyAuth
func (h *TaskHandler) create(c echo.Context) error {
	form, err := newTaskForm(c)
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

	projectOId, err := primitive.ObjectIDFromHex(form.ProjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	task := repository.Task{
		ID:          primitive.NewObjectID(),
		Name:        form.Name,
		Description: form.Description,
		StartDate:   time.Unix(form.StartDate, 0),
		EndDate:     time.Unix(form.EndDate, 0),
		Contributor: contributorsOId,
		Status:      _const.TaskActive,
		ProjectID:   projectOId,
		CreatedAt:   time.Now(),
		IsDeleted:   false,
	}

	err = h.taskRepo.CreateOne(&task)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, task)
}

// Delete Task
// @Tags Task
// @Summary Delete task
// @ID task-delete
// @Router /api/task/{id} [delete]
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200
// @Security ApiKeyAuth
func (h *TaskHandler) delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Task ID cannot be empty.")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	err = h.taskRepo.DeleteOneByID(objectID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, "Task deleted")
}
