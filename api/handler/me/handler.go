package me

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/config"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/file"
	"proman-backend/internal/pkg/log"
	_mongo "proman-backend/internal/pkg/mongo"
	"proman-backend/internal/pkg/util"
)

type Handler struct {
	userRepo     *repository.UserCollRepository
	projectRepo  *repository.ProjectCollRepository
	taskRepo     *repository.TaskCollRepository
	scheduleRepo *repository.ScheduleCollRepository
	codeRepo     *repository.CodeCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		userRepo:     repository.NewUserCollRepository(db),
		projectRepo:  repository.NewProjectCollRepository(db),
		taskRepo:     repository.NewTaskCollRepository(db),
		scheduleRepo: repository.NewScheduleCollRepository(db),
		codeRepo:     repository.NewCodeCollRepository(db),
	}

	me := e.Group("/api", context.ContextHandler)

	me.GET("/me", h.myProfile)
	me.PUT("/me", h.updateMyProfile)
	me.PUT("/me/password", h.updateMyPassword)

	me.GET("/me/schedules", h.mySchedule)

	me.GET("/me/projects", h.myProjects)
	me.GET("/me/project/count", h.myProjectCount)
	me.GET("/me/project/count/type", h.myProjectCountByType)

	me.GET("/me/tasks", h.myTasks)
	me.GET("/me/task/count", h.myTaskCount)
	me.GET("/me/task/overview", h.myTaskOverview)
	me.GET("/me/task/status", h.myTaskStatus)

	return h
}

// My Profile
// @Tags Me
// @Summary Get my info
// @ID my-profile
// @Router /api/me [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myProfile(c echo.Context) error {
	uc := c.(*context.Context)

	user, err := h.userRepo.FindOneByID(uc.Claims.IDAsObjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, user)
}

// Update My Profile
// @Tags Me
// @Summary Update my profile
// @ID update-my-profile
// @Router /api/me [put]
// @Accept json
// @Param name formData string false "name"
// @Param email formData string false "email"
// @Param position formData string false "position"
// @Param phone formData string false "phone"
// @Param avatar formData file false "avatar"
// @Param verification_code formData string true "verification_code"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) updateMyProfile(c echo.Context) error {
	uc := c.(*context.Context)
	docForm, err := newUpdateMyProfileForm(c)
	if err != nil {
		return err
	}

	user, err := h.userRepo.FindOneByID(uc.Claims.IDAsObjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	if config.Vcode.CheckEnable {
		code, err := h.codeRepo.FindActiveOneByUserID(uc.Claims.IDAsObjectID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid verification code")
			}
			log.Errorf("Error finding code: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
		}

		if code == nil || code.Code != docForm.VerificationCode {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid verification code")
		}
	}

	avatar, _ := file.GetFileThenUpload(c, "avatar", config.AWS.AvatarDir)
	if avatar != "" {
		user.Avatar = avatar
	}

	if docForm.Name != "" {
		user.Name = docForm.Name
	}
	if docForm.Email != "" {
		user.Email = docForm.Email
	}
	if docForm.Position != "" {
		user.Position = docForm.Position
	}
	if docForm.Phone != "" {
		user.Phone = docForm.Phone
	}

	doc, err := h.userRepo.Update(user)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, doc)
}

// Update My Password
// @Tags Me
// @Summary Update my password
// @ID update-my-password
// @Router /api/me/password [put]
// @Accept json
// @Param body body updateMyPasswordForm true "update my password json"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) updateMyPassword(c echo.Context) error {
	uc := c.(*context.Context)
	docForm, err := newUpdateMyPasswordForm(c)
	if err != nil {
		return err
	}

	user, err := h.userRepo.FindOneByID(uc.Claims.IDAsObjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	if !util.CheckPassword(user.Password, docForm.OldPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, "Wrong old password")
	}

	if config.Vcode.CheckEnable {
		code, err := h.codeRepo.FindActiveOneByUserID(uc.Claims.IDAsObjectID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid verification code")
			}
			log.Errorf("Error finding code: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
		}

		if code == nil || code.Code != docForm.VerificationCode {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid verification code")
		}
	}

	user.Password = util.CryptPassword(docForm.NewPassword)

	doc, err := h.userRepo.Update(user)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, doc)
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

	cq := util.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	schedules, err := h.scheduleRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Schedule not found")
		}
		log.Errorf("Error finding schedule: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	response := make([]map[string]interface{}, 0)
	contributors := map[bson.ObjectID]string{}

	for _, schedule := range schedules {
		contributorInList := make([]string, 0)

		for _, contributor := range schedule.Contributor {
			if name, exists := contributors[contributor]; exists {
				contributorInList = append(contributorInList, name)
			} else if user, err := h.userRepo.FindOneByID(contributor); err == nil {
				contributors[contributor] = user.Name
				contributorInList = append(contributorInList, user.Name)
			}
		}

		response = append(response, map[string]interface{}{
			"id":          schedule.ID,
			"name":        schedule.Name,
			"description": schedule.Description,
			"start_date":  schedule.StartDate,
			"end_date":    schedule.EndDate,
			"start_time":  schedule.StartTime,
			"end_time":    schedule.EndTime,
			"contributor": contributorInList,
			"type":        schedule.Type,
			"created_at":  schedule.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, response)
}

// My Projects
// @Tags Me Project
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

	cq := util.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

	limit := cq.Limit
	page := cq.Page

	projects, err := h.projectRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Project not found")
		}
		log.Errorf("Error finding project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	cq.ResetPagination()
	totalProjects, err := h.projectRepo.CountProject(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Project not found")
		}
		log.Errorf("Error counting project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	result := _mongo.MakePaginateResult(projects, int64(totalProjects.Total), page, limit)
	return c.JSON(http.StatusOK, result)
}

// My Project Count
// @Tags Me Project
// @Summary Get my project count
// @ID my-project-count
// @Router /api/me/project/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myProjectCount(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util.NewCommonQuery(c)
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
// @Tags Me Project
// @Summary Get my project count by type
// @ID my-project-count-type
// @Router /api/me/project/count/type [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myProjectCountByType(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util.NewCommonQuery(c)
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
// @Tags Me Task
// @Summary Get my tasks
// @ID my-tasks
// @Router /api/me/tasks [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTasks(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util.NewCommonQuery(c)
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
// @Tags Me Task
// @Summary Get my task count
// @ID my-task-count
// @Router /api/me/task/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTaskCount(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

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
		return c.JSON(http.StatusOK, repository.CountTaskDetail{})
	}
}

// My Task Overview
// @Tags Me Task
// @Summary Get my task overview
// @ID my-task-overview
// @Router /api/me/task/overview [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTaskOverview(c echo.Context) error {
	doc := []repository.TaskOverview{}
	uc := c.(*context.Context)

	cq := util.NewCommonQuery(c)
	cq.UserId = uc.Claims.IDAsObjectID

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

// My Task List By Status
// @Tags Me Task
// @Summary Get my task list by status
// @ID my-task-list-status
// @Router /api/me/task/status [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) myTaskStatus(c echo.Context) error {
	uc := c.(*context.Context)

	cq := util.NewCommonQuery(c)
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
