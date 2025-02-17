package user

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/config"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/log"
	_mongo "proman-backend/internal/pkg/mongo"
	"proman-backend/internal/pkg/util"
	"strings"
)

type Handler struct {
	userRepo    *repository.UserCollRepository
	taskRepo    *repository.TaskCollRepository
	projectRepo *repository.ProjectCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		userRepo:    repository.NewUserCollRepository(db),
		taskRepo:    repository.NewTaskCollRepository(db),
		projectRepo: repository.NewProjectCollRepository(db),
	}

	user := e.Group("/api", context.ContextHandler)

	user.GET("/users", h.userList)
	user.GET("/user/count", h.userCount)

	return h
}

// User Count
// @Tags User
// @Summary Get user count
// @ID user-count
// @Router /api/user/count [get]
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) userCount(c echo.Context) error {
	count, err := h.taskRepo.CountUserTask(h.userRepo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		log.Errorf("Error counting user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, count)
}

// User List
// @Tags User
// @Summary Get list users
// @ID user-list
// @Router /api/users [get]
// @Param q query string false "Search by nama"
// @Param sort query string false "Sort" enums(asc,desc)
// @Param page query int false "Page number pagination"
// @Param limit query int false "Limit pagination"
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
func (h *Handler) userList(c echo.Context) error {
	cq := util.NewCommonQuery(c)

	limit := cq.Limit
	page := cq.Page

	users, err := h.userRepo.FindAllUsers(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	for i, user := range users {
		if _, ok := user["avatar"].(string); ok && user["avatar"] != "" {
			url := "https://" + config.S3.Bucket
			if !strings.Contains(config.S3.EndPoint, "https://") {
				users[i]["avatar"] = url + "." + config.S3.EndPoint + "/" + users[i]["avatar"].(string)
			} else {
				users[i]["avatar"] = url + "." + config.S3.EndPoint[8:] + "/" + users[i]["avatar"].(string)
			}
		}
	}

	cq.ResetPagination()
	totalUsers, err := h.userRepo.FindAllUsers(cq)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	result := _mongo.MakePaginateResult(users, int64(len(totalUsers)), page, limit)
	return c.JSON(http.StatusOK, result)
}
