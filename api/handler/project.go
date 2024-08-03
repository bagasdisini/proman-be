package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/config"
	_const "proman-backend/pkg/const"
	"proman-backend/pkg/context"
	"proman-backend/pkg/file"
	"strings"
	"time"
)

type projectForm struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	StartDate   int64  `json:"start_date" form:"start_date"`
	EndDate     int64  `json:"end_date" form:"end_date"`
	Contributor string `json:"contributor" form:"contributor"`
	Type        string `json:"type" form:"type"`
	Logo        string `json:"logo" bson:"logo"`
	Attachments string `json:"attachments" bson:"attachments"`
}

func newProjectForm(c echo.Context) (*projectForm, error) {
	form := projectForm{}
	if err := c.Bind(&form); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid data format.")
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Description = strings.TrimSpace(form.Description)
	form.Contributor = strings.TrimSpace(form.Contributor)
	form.Type = strings.TrimSpace(form.Type)

	if form.Name == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Name cannot be empty.")
	}
	if form.Description == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Description cannot be empty.")
	}
	if form.Contributor == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Contributor cannot be empty.")
	}
	if form.Type == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Type cannot be empty.")
	}
	if form.StartDate <= 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid start date.")
	}
	if form.EndDate <= 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid end date.")
	}
	return &form, nil
}

type ProjectHandler struct {
	projectRepo *repository.ProjectCollRepository
	taskRepo    *repository.TaskCollRepository
}

func NewProjectHandler(e *echo.Echo, db *mongo.Database) *ProjectHandler {
	h := &ProjectHandler{
		projectRepo: repository.NewProjectRepository(db),
		taskRepo:    repository.NewTaskRepository(db),
	}

	project := e.Group("/api", context.ContextHandler)

	project.GET("/project", h.list)
	project.GET("/project/:id", h.detail)

	project.GET("/project/count", h.count)
	project.GET("/project/count/type", h.countByType)

	project.GET("/project/user/:id", h.listByUser)

	project.POST("/project", h.create)

	project.DELETE("/project/:id", h.delete)

	return h
}

// List Project
// @Tags Project
// @Summary Get list of project
// @ID list-project
// @Router /api/project [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) list(c echo.Context) error {
	projects, err := h.projectRepo.FindAll(h.taskRepo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, projects)
}

// Get Project
// @Tags Project
// @Summary Get project by id
// @ID get-project
// @Router /api/project/{id} [get]
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) detail(c echo.Context) error {
	id := c.Param("id")

	OId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	project, err := h.projectRepo.FindOneByID(OId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, project)
}

// Project Count
// @Tags Project
// @Summary Get project count
// @ID project-count
// @Router /api/project/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) count(c echo.Context) error {
	currentEnd := time.Date(time.Now().Year(), 12, 31, 23, 59, 59, 0, time.Local)
	currentStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Local)
	current, err := h.projectRepo.CountProject(currentStart, currentEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}

	prevEnd := currentStart.AddDate(0, 0, -1)
	prevStart := time.Date(prevEnd.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	previous, err := h.projectRepo.CountProject(prevStart, prevEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}

	docs := repository.CountProject{}
	if len(*current) > 0 {
		docs.Current = (*current)[0]
	} else {
		docs.Current = repository.CountProjectDetail{}
	}
	if len(*previous) > 0 {
		docs.LastYear = (*previous)[0]
	} else {
		docs.LastYear = repository.CountProjectDetail{}
	}
	return c.JSON(http.StatusOK, docs)
}

// Project Count By Type
// @Tags Project
// @Summary Get project count by type
// @ID project-count-type
// @Router /api/project/count/type [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) countByType(c echo.Context) error {
	count, err := h.projectRepo.CountTypeProject()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, count)
}

// List Project By User
// @Tags Project
// @Summary Get list of project by user
// @ID list-project-user
// @Router /api/project/user/{id} [get]
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) listByUser(c echo.Context) error {
	id := c.Param("id")

	OId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID.")
	}

	projects, err := h.projectRepo.FindAllByContributorID(OId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, projects)
}

// Create Project
// @Tags Project
// @Summary Create project
// @ID create-project
// @Router /api/project [post]
// @Accept json
// @Produce json
// @Param name formData string true "Project name"
// @Param description formData string true "Project description"
// @Param start_date formData int true "Project start date"
// @Param end_date formData int true "Project end date"
// @Param contributor formData string true "Project contributor"
// @Param type formData string true "Project type"
// @Param logo formData file false "Project logo"
// @Param attachments formData file false "Project attachments"
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) create(c echo.Context) error {
	form, err := newProjectForm(c)
	if err != nil {
		return err
	}

	tokenData := c.(*context.Context)

	ownerIncluded := false
	contributorsOId := make([]primitive.ObjectID, 0)
	for _, user := range strings.Split(form.Contributor, ",") {
		userOId, err := primitive.ObjectIDFromHex(user)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor.")
		}
		if userOId == tokenData.Claims.IDAsObjectID {
			ownerIncluded = true
		}
		contributorsOId = append(contributorsOId, userOId)
	}
	if !ownerIncluded {
		contributorsOId = append(contributorsOId, tokenData.Claims.IDAsObjectID)
	}

	logo, err := file.GetFileThenUpload(c, "logo", config.AWS.ProjectLogoDir)
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return err
	}
	attachments, err := file.GetFilesThenUpload(c, "attachments", config.AWS.FileDir)
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return err
	}

	project := repository.Project{
		ID:          primitive.NewObjectID(),
		Name:        form.Name,
		Description: form.Description,
		Type:        form.Type,
		StartDate:   time.UnixMilli(form.StartDate),
		EndDate:     time.UnixMilli(form.EndDate),
		Contributor: contributorsOId,
		Status:      _const.ProjectActive,
		Attachments: attachments,
		Logo:        logo,
		CreatedAt:   time.Now(),
		IsDeleted:   false,
	}

	doc, err := h.projectRepo.InsertOne(&project)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to create project.")
	}
	return c.JSON(http.StatusOK, doc)
}

// Delete Project
// @Tags Project
// @Summary Delete project by id
// @ID delete-project
// @Router /api/project/{id} [delete]
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) delete(c echo.Context) error {
	id := c.Param("id")

	OId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	project, err := h.projectRepo.FindOneByID(OId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}

	err = h.projectRepo.DeleteOneByID(project.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete project.")
	}

	err = h.taskRepo.DeleteAllByProjectID(project.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete project.")
	}
	return c.JSON(http.StatusOK, "Project deleted.")
}
