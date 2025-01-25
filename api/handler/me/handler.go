package me

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/log"
	_mongo "proman-backend/internal/pkg/mongo"
	util2 "proman-backend/internal/pkg/util"
)

type Handler struct {
	userRepo     *repository.UserCollRepository
	projectRepo  *repository.ProjectCollRepository
	taskRepo     *repository.TaskCollRepository
	scheduleRepo *repository.ScheduleCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		userRepo:     repository.NewUserRepository(db),
		projectRepo:  repository.NewProjectRepository(db),
		taskRepo:     repository.NewTaskRepository(db),
		scheduleRepo: repository.NewScheduleRepository(db),
	}

	me := e.Group("/api", context.ContextHandler)

	me.GET("/me", h.me)

	me.GET("/me/projects", h.myProjects)
	me.GET("/me/project/count", h.myProjectCount)
	me.GET("/me/project/count/type", h.myProjectCountByType)

	me.GET("/me/tasks", h.myTasks)
	me.GET("/me/task/count", h.myTaskCount)
	me.GET("/me/task/overview", h.myTaskOverview)
	me.GET("/me/task/status", h.myTaskStatus)

	me.GET("/me/schedules", h.mySchedule)

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
func (h *Handler) me(c echo.Context) error {
	uc := c.(*context.Context)

	user, err := h.userRepo.FindOneByID(uc.Claims.IDAsObjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
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
// @Router /api/me/schedules [get]
// @Param q query string false "Search by name"
// @Param type query string false "Search by type" Enums(all, meeting, discussion, review, presentation, etc)
// @Param start query string false "Start date"
// @Param end query string false "End date"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) mySchedule(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	schedules, err := h.scheduleRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Schedule not found")
		}
		log.Errorf("Error finding schedule: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, schedules)
}

// My Projects
// @Tags Me
// @Summary Get my projects
// @ID my-projects
// @Router /api/me/projects [get]
// @Param q query string false "Search by nama or description"
// @Param status query string false "Search by status" Enums(active, completed, pending, cancelled)
// @Param start query string false "Start date"
// @Param end query string false "End date"
// @Param sort query string false "Sort" enums(asc,desc)
// @Param page query int false "Page number pagination"
// @Param limit query int false "Limit pagination"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myProjects(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	projects, err := h.projectRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Project not found")
		}
		log.Errorf("Error finding project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	cq.ResetDate()
	totalProjects, err := h.projectRepo.CountProject(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Project not found")
		}
		log.Errorf("Error counting project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	result := _mongo.MakeResult(projects, int64(totalProjects.Total), cq.Page, cq.Limit)
	return c.JSON(http.StatusOK, result)
}

// My Project Count
// @Tags Me
// @Summary Get my project count
// @ID my-project-count
// @Router /api/me/project/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myProjectCount(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	count, err := h.projectRepo.CountProject(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Project not found")
		}
		log.Errorf("Error counting project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	return c.JSON(http.StatusOK, count)
}

// My Project Count By Type
// @Tags Me
// @Summary Get my project count by type
// @ID my-project-count-type
// @Router /api/me/project/count/type [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myProjectCountByType(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	count, err := h.projectRepo.CountProjectTypes(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Project not found")
		}
		log.Errorf("Error counting project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, count)
}

// My Tasks
// @Tags Me
// @Summary Get my tasks
// @ID my-tasks
// @Router /api/me/tasks [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTasks(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

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

// My Task Count
// @Tags Me
// @Summary Get my task count
// @ID my-task-count
// @Router /api/me/task/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTaskCount(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	count, err := h.taskRepo.CountTask(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Task not found")
		}
		log.Errorf("Error counting task: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	if len(*count) > 0 {
		return c.JSON(http.StatusOK, (*count)[0])
	} else {
		return c.JSON(http.StatusOK, repository.CountProjectDetail{})
	}
}

// My Task Overview
// @Tags Me
// @Summary Get my task overview
// @ID my-task-overview
// @Router /api/me/task/overview [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTaskOverview(c echo.Context) error {
	var doc []repository.TaskOverview
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	for _, v := range []int{-7, -6, -5, -4, -3, -2, -1, 0} {
		cq.Start = util2.StartOfWeek(v)
		cq.End = util2.EndOfWeek(v)

		count, err := h.taskRepo.CountTask(cq)
		if err != nil || len(*count) == 0 {
			doc = append(doc, repository.TaskOverview{
				Start: cq.Start.Format("02 Jan"),
				End:   cq.End.Format("02 Jan"),
				Count: 0,
			})
			log.Warnf("Error counting task: %v, count: %v", err, count)
			continue
		}

		doc = append(doc, repository.TaskOverview{
			Start: cq.Start.Format("02 Jan"),
			End:   cq.End.Format("02 Jan"),
			Count: (*count)[0].Active + (*count)[0].Testing + (*count)[0].Completed,
		})
	}
	return c.JSON(http.StatusOK, doc)
}

// My Task List By Status
// @Tags Me
// @Summary Get my task list by status
// @ID my-task-list-status
// @Router /api/me/task/status [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTaskStatus(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util2.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

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
