package task

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/log"
	"proman-backend/internal/pkg/util"
	"strings"
	"time"
)

type Handler struct {
	taskRepo *repository.TaskCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		taskRepo: repository.NewTaskCollRepository(db),
	}

	task := e.Group("/api", context.ContextHandler)

	task.GET("/task/:id", h.task)
	task.GET("/tasks", h.tasks)
	task.GET("/task/count", h.count)
	task.GET("/task/overview", h.overview)
	task.GET("/task/status", h.status)

	task.POST("/task", h.create)

	task.PUT("/task/:id", h.update)

	task.DELETE("/task/:id", h.delete)

	return h
}

// Task
// @Tags Task
// @Summary Get task by ID
// @ID task
// @Router /api/task/{id} [get]
// @Param id path string true "Task ID"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) task(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Task ID cannot be empty.")
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	task, err := h.taskRepo.FindOneByID(objectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Task not found")
		}
		log.Errorf("Error finding task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, task)
}

// Tasks
// @Tags Task
// @Summary Get tasks
// @ID tasks
// @Router /api/tasks [get]
// @Param q query string false "Search by nama or description"
// @Param status query string false "Search by status" Enums(active, testing, completed, cancelled)
// @Param userId query string false "Search by contributor"
// @Param projectId query string false "Search by project"
// @Param start query string false "Start date"
// @Param end query string false "End date"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) tasks(c echo.Context) error {
	cq := util.NewCommonQuery(c)
	tasks, err := h.taskRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Task not found")
		}
		log.Errorf("Error finding task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, tasks)
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
func (h *Handler) count(c echo.Context) error {
	cq := util.NewCommonQuery(c)
	count, err := h.taskRepo.CountTask(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Task not found")
		}
		log.Errorf("Error counting task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	if len(count) > 0 {
		return c.JSON(http.StatusOK, count[0])
	} else {
		return c.JSON(http.StatusOK, repository.CountProjectDetail{})
	}
}

// Task Overview
// @Tags Task
// @Summary Get task overview
// @ID task-overview
// @Router /api/task/overview [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) overview(c echo.Context) error {
	doc := []repository.TaskOverview{}
	cq := util.NewCommonQuery(c)
	for _, v := range []int{-7, -6, -5, -4, -3, -2, -1, 0} {
		cq.Start = util.StartOfWeek(v)
		cq.End = util.EndOfWeek(v)

		count, err := h.taskRepo.CountTask(cq)
		if err != nil || len(count) == 0 {
			doc = append(doc, repository.TaskOverview{
				Start: cq.Start.Format("02 Jan"),
				End:   cq.End.Format("02 Jan"),
				Count: 0,
			})
			continue
		}

		doc = append(doc, repository.TaskOverview{
			Start: cq.Start.Format("02 Jan"),
			End:   cq.End.Format("02 Jan"),
			Count: count[0].Active + count[0].Testing + count[0].Completed,
		})
	}
	return c.JSON(http.StatusOK, doc)
}

// Task By Status
// @Tags Task
// @Summary Get task list by status
// @ID task-list-status
// @Router /api/task/status [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) status(c echo.Context) error {
	cq := util.NewCommonQuery(c)

	docs := repository.TaskGroup{}

	cq.Status = _const.TaskActive
	active, err := h.taskRepo.FindAll(cq)
	if err == nil {
		docs.Active = active
	} else {
		log.Warnf("Error finding active task: %v", err)
	}

	cq.Status = _const.TaskTesting
	testing, err := h.taskRepo.FindAll(cq)
	if err == nil {
		docs.Testing = testing
	} else {
		log.Warnf("Error finding testing task: %v", err)
	}

	cq.Status = _const.TaskCompleted
	completed, err := h.taskRepo.FindAll(cq)
	if err == nil {
		docs.Completed = completed
	} else {
		log.Warnf("Error finding completed task: %v", err)
	}

	cq.Status = _const.TaskCancelled
	cancelled, err := h.taskRepo.FindAll(cq)
	if err == nil {
		docs.Cancelled = cancelled
	} else {
		log.Warnf("Error finding cancelled task: %v", err)
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
func (h *Handler) create(c echo.Context) error {
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

// Update Task
// @Tags Task
// @Summary Update task
// @ID task-update
// @Router /api/task/{id} [put]
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body updateTaskForm true "Update task data"
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.")
	}

	form, err := newUpdateTaskForm(c)
	if err != nil {
		return err
	}

	contributorsOId := make([]bson.ObjectID, 0)
	if len(form.Contributor) != 0 {
		for _, user := range strings.Split(form.Contributor, ",") {
			userOId, err := bson.ObjectIDFromHex(user)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor.")
			}
			contributorsOId = append(contributorsOId, userOId)
		}
	}

	task, err := h.taskRepo.FindOneByID(objectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Task not found")
		}
		log.Errorf("Error finding task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	if len(form.Name) != 0 {
		task.Name = form.Name
	}
	if len(form.Description) != 0 {
		task.Description = form.Description
	}
	if form.StartDate != 0 {
		task.StartDate = time.UnixMilli(form.StartDate)
	}
	if form.EndDate != 0 {
		task.EndDate = time.UnixMilli(form.EndDate)
	}
	if len(form.Contributor) != 0 {
		task.Contributor = contributorsOId
	}
	if len(form.Status) != 0 {
		task.Status = form.Status
	}

	err = h.taskRepo.UpdateOneByID(task)
	if err != nil {
		log.Errorf("Failed to update task: %v", err)
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
func (h *Handler) delete(c echo.Context) error {
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
