package option

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/config"
	_const "proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/util"
)

type Handler struct {
	userRepo    *repository.UserCollRepository
	projectRepo *repository.ProjectCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		userRepo:    repository.NewUserCollRepository(db),
		projectRepo: repository.NewProjectCollRepository(db),
	}

	option := e.Group("/api", middleware.BasicAuth(
		func(username, password string, c echo.Context) (bool, error) {
			return username == config.Basic.Username && password == config.Basic.Password, nil
		}))

	option.GET("/option/type/position", h.position)
	option.GET("/option/type/project", h.projectType)
	option.GET("/option/type/schedule", h.scheduleType)

	option.GET("/option/user", h.user)
	option.GET("/option/project", h.project)

	return h
}

// Get Position
// @Tags Option
// @Summary Get position
// @ID option-position
// @Router /api/option/type/position [get]
// @Accept json
// @Produce json
// @Success 200
// @Security BasicAuth
func (h *Handler) position(c echo.Context) error {
	return c.JSON(http.StatusOK, _const.GetAllPositions())
}

// Get Project Type
// @Tags Option
// @Summary Get project type
// @ID option-project-type
// @Router /api/option/type/project [get]
// @Accept json
// @Produce json
// @Success 200
// @Security BasicAuth
func (h *Handler) projectType(c echo.Context) error {
	return c.JSON(http.StatusOK, _const.GetAllProjectTypes())
}

// Get Schedule Type
// @Tags Option
// @Summary Get schedule type
// @ID option-schedule-type
// @Router /api/option/type/schedule [get]
// @Accept json
// @Produce json
// @Success 200
// @Security BasicAuth
func (h *Handler) scheduleType(c echo.Context) error {
	return c.JSON(http.StatusOK, _const.GetAllScheduleTypes())
}

// Get User
// @Tags Option
// @Summary Get user
// @ID option-user
// @Router /api/option/user [get]
// @Accept json
// @Produce json
// @Success 200
// @Security BasicAuth
func (h *Handler) user(c echo.Context) error {
	response := make([]map[string]interface{}, 0)

	cq := util.NilCommonQuery()

	usersDoc, err := h.userRepo.FindAllUsers(cq)
	if err != nil {
		return c.JSON(http.StatusOK, response)
	}

	for _, user := range usersDoc {
		response = append(response, map[string]interface{}{
			"_id":  user["_id"],
			"name": user["name"],
		})
	}
	return c.JSON(http.StatusOK, response)
}

// Get Project
// @Tags Option
// @Summary Get project
// @ID option-project
// @Router /api/option/project [get]
// @Accept json
// @Produce json
// @Success 200
// @Security BasicAuth
func (h *Handler) project(c echo.Context) error {
	response := make([]map[string]interface{}, 0)

	cq := util.NilCommonQuery()

	projectsDoc, err := h.projectRepo.FindAll(cq)
	if err != nil {
		return c.JSON(http.StatusOK, response)
	}

	for _, project := range projectsDoc {
		response = append(response, map[string]interface{}{
			"_id":  project.ID,
			"name": project.Name,
		})
	}
	return c.JSON(http.StatusOK, response)
}
