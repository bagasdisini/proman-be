package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/pkg/context"
)

type MeHandler struct {
	userRepo *repository.UserCollRepository
}

func NewMeHandler(e *echo.Echo, db *mongo.Database) *MeHandler {
	h := &MeHandler{
		userRepo: repository.NewUserRepository(db),
	}

	me := e.Group("/api", context.ContextHandler)

	me.GET("/me", h.me)

	return h
}

// Me
// @Tags Me
// @Summary Get my info
// @ID me
// @Router /api/me [get]
// @Accept json
// @Produce  json
// @Success 200
// @Security ApiKeyAuth
func (h *MeHandler) me(c echo.Context) error {
	uc := c.(*context.Context)

	user, err := h.userRepo.FindOneByID(uc.Claims.IDAsObjectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}
		return err
	}

	if user.IsDeleted {
		return echo.NewHTTPError(http.StatusUnauthorized, "User is deleted")
	}

	return c.JSON(http.StatusOK, user)
}
