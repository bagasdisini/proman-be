package project

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
	"proman-backend/internal/pkg/util"
	"strings"
	"time"
)

type Handler struct {
	projectRepo *repository.ProjectCollRepository
	taskRepo    *repository.TaskCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		projectRepo: repository.NewProjectCollRepository(db),
		taskRepo:    repository.NewTaskCollRepository(db),
	}

	project := e.Group("/api", context.ContextHandler)

	project.GET("/projects", h.list)
	project.GET("/project/:id", h.detail)

	project.POST("/project", h.create)

	project.DELETE("/project/:id", h.delete)

	return h
}

// List Project
// @Tags Project
// @Summary Get list of project
// @ID list-project
// @Router /api/projects [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) list(c echo.Context) error {
	cq := util.NewCommonQuery(c)

	projects, err := h.projectRepo.FindAll(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		log.Errorf("Error finding project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
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
func (h *Handler) detail(c echo.Context) error {
	id := c.Param("id")

	oId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	project, err := h.projectRepo.FindOneByID(oId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		log.Errorf("Error finding project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, project)
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
func (h *Handler) create(c echo.Context) error {
	form, err := newProjectForm(c)
	if err != nil {
		return err
	}

	tokenData := c.(*context.Context)

	ownerIncluded := false
	contributorsOId := make([]bson.ObjectID, 0)
	for _, user := range strings.Split(form.Contributor, ",") {
		userOId, err := bson.ObjectIDFromHex(user)
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
		log.Errorf("Failed to upload logo: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	attachments, err := file.GetFilesThenUpload(c, "attachments", config.AWS.FileDir)
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		log.Errorf("Failed to upload attachments: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	project := repository.Project{
		ID:          bson.NewObjectID(),
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
		log.Errorf("Failed to create project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
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
func (h *Handler) delete(c echo.Context) error {
	id := c.Param("id")

	oId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.")
	}

	project, err := h.projectRepo.FindOneByID(oId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		log.Errorf("Error finding project: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	err = h.projectRepo.DeleteOneByID(project.ID)
	if err != nil {
		log.Errorf("Failed to delete project: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "There was an error, please try again")
	}

	err = h.taskRepo.DeleteAllByProjectID(project.ID)
	if err != nil {
		log.Errorf("Failed to delete project: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, "Project deleted.")
}
