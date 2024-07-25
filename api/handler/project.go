package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/config"
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
}

func NewProjectHandler(e *echo.Echo, db *mongo.Database) *ProjectHandler {
	h := &ProjectHandler{
		projectRepo: repository.NewProjectRepository(db),
	}

	project := e.Group("/api", context.ContextHandler)

	project.GET("/project/count", h.projectCount)
	project.GET("/project/count/type", h.projectCountByType)

	project.POST("/project", h.createProject)

	return h
}

// Project Count
// @Tags Project
// @Summary Get project count
// @ID project-count
// @Router /api/project/count [get]
// @Accept json
// @Produce  json
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) projectCount(c echo.Context) error {
	currentEnd := time.Now()
	currentStart := currentEnd.AddDate(0, 0, -30)
	current, err := h.projectRepo.CountProject(currentStart, currentEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}

	prevEnd := currentStart.Add(-time.Second)
	prevStart := prevEnd.AddDate(0, 0, -30)
	previous, err := h.projectRepo.CountProject(prevStart, prevEnd)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}

	docs := repository.CountProject{
		Current:   *current,
		LastMonth: *previous,
	}
	return c.JSON(http.StatusOK, docs)
}

// Project Count By Type
// @Tags Project
// @Summary Get project count by type
// @ID project-count-type
// @Router /api/project/count/type [get]
// @Accept json
// @Produce  json
// @Success 200
// @Security ApiKeyAuth
func (h *ProjectHandler) projectCountByType(c echo.Context) error {
	count, err := h.projectRepo.CountTypeProject()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, count)
}

// Create Project
// @Tags Project
// @Summary Create project
// @ID create-project
// @Router /api/project [post]
// @Accept json
// @Produce  json
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
func (h *ProjectHandler) createProject(c echo.Context) error {
	form, err := newProjectForm(c)
	if err != nil {
		return err
	}

	tokenData := c.Get("me").(*context.UserClaims)

	ownerIncluded := false
	contributorsOId := make([]primitive.ObjectID, 0)
	for _, user := range strings.Split(form.Contributor, ",") {
		userOId, err := primitive.ObjectIDFromHex(user)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid contributor.")
		}
		if userOId == tokenData.IDAsObjectID {
			ownerIncluded = true
		}
		contributorsOId = append(contributorsOId, userOId)
	}
	if !ownerIncluded {
		contributorsOId = append(contributorsOId, tokenData.IDAsObjectID)
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
		StartDate:   time.Unix(form.StartDate, 0),
		EndDate:     time.Unix(form.EndDate, 0),
		Contributor: contributorsOId,
		Status:      "active",
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
