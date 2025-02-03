package code

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"net/url"
	"proman-backend/api/repository"
	"proman-backend/config"
	"proman-backend/internal/pkg/log"
	"proman-backend/internal/pkg/mail"
	"proman-backend/internal/pkg/util"
	"time"
)

type Handler struct {
	userRepo *repository.UserCollRepository
	codeRepo *repository.CodeCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		userRepo: repository.NewUserRepository(db),
		codeRepo: repository.NewCodeCollRepository(db),
	}

	group := e.Group("/api", middleware.BasicAuth(
		func(username, password string, c echo.Context) (bool, error) {
			return username == config.Basic.Username && password == config.Basic.Password, nil
		}))

	group.POST("/verification-code/:email", h.vcode)

	return h
}

// Verification Code
// @Tags Code
// @Summary Create verification code
// @ID code-create
// @Router /api/verification-code/{email} [post]
// @Accept json
// @Produce json
// @Param email path string true "Email User"
// @Success 200
// @Security BasicAuth
func (h *Handler) vcode(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email cannot be empty.")
	}
	email, err := url.QueryUnescape(email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid email format.")
	}

	user, err := h.userRepo.FindOneByEmail(email)
	if err != nil {
		log.Errorf("Error finding user by email: %v", err)
		return c.JSON(http.StatusOK, map[string]interface{}{"message": "Success, please check your email."})
	}

	codeDoc := repository.Code{
		ID:        bson.NewObjectID(),
		UserID:    user.ID,
		Email:     user.Email,
		Code:      util.RandomString(config.Vcode.Length),
		IsUsed:    false,
		ExpiredAt: time.Now().Add(time.Minute * 15),
		CreatedAt: time.Now(),
	}

	_, err = h.codeRepo.InsertOne(&codeDoc)
	if err != nil {
		log.Errorf("Error inserting code: %v", err)
		return c.JSON(http.StatusOK, map[string]interface{}{"message": "Success, please check your email."})
	}

	go func(email string, template string) {
		err = mail.SendMail(nil, []string{email}, "Verification Code", template)
		if err != nil {
			log.Errorf("Error sending email: %v", err)
			return
		}
	}(user.Email, mail.VerificationCode(codeDoc.Code))
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Success, please check your email."})
}
