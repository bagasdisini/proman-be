package handler

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/log"
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
		log.Errorf("Error binding task form: %v", err)
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
	_, err := bson.ObjectIDFromHex(form.ProjectID)
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

	task.POST("/task", h.create)

	task.DELETE("/task/:id", h.delete)

	return h
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

	contributorsOId := make([]bson.ObjectID, 0)
	for _, user := range strings.Split(form.Contributor, ",") {
		userOId, err := bson.ObjectIDFromHex(user)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor.")
		}
		contributorsOId = append(contributorsOId, userOId)
	}

	projectOId, err := bson.ObjectIDFromHex(form.ProjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	task := repository.Task{
		ID:          bson.NewObjectID(),
		Name:        form.Name,
		Description: form.Description,
		StartDate:   time.UnixMilli(form.StartDate),
		EndDate:     time.UnixMilli(form.EndDate),
		Contributor: contributorsOId,
		Status:      _const.TaskActive,
		ProjectID:   projectOId,
		CreatedAt:   time.Now(),
		IsDeleted:   false,
	}

	err = h.taskRepo.CreateOne(&task)
	if err != nil {
		log.Errorf("Failed to create task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
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

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	err = h.taskRepo.DeleteOneByID(objectID)
	if err != nil {
		log.Errorf("Failed to delete task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, "Task deleted")
}
