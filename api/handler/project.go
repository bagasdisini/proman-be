package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
)

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
	count, err := h.projectRepo.CountProject()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Project not found")
		}
		return err
	}
	return c.JSON(http.StatusOK, count)
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
