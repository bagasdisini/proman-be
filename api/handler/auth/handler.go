package auth

import (
	"errors"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"net/http"
	"proman-backend/api/repository"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/context"
	"proman-backend/internal/pkg/log"
	"proman-backend/internal/pkg/mail"
	"proman-backend/internal/pkg/util"
	"time"
)

type Handler struct {
	userRepo *repository.UserCollRepository
}

func NewHandler(e *echo.Echo, db *mongo.Database) *Handler {
	h := &Handler{
		userRepo: repository.NewUserRepository(db),
	}

	e.POST("/api/login", h.login)
	e.POST("/api/register", h.register)

	e.POST("/api/forgot-password", h.forgotPassword)

	return h
}

// Login
// @Tags Auth
// @Summary Login
// @ID login
// @Router /api/login [post]
// @Accept json
// @Param body body loginForm true "login json"
// @Produce json
// @Success 200
func (h *Handler) login(c echo.Context) error {
	loginForm, err := newLoginForm(c)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindOneByEmail(loginForm.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Wrong username/email or password")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	if !util.CheckPassword(u.Password, loginForm.Password) {
		return echo.NewHTTPError(http.StatusBadRequest, "Wrong username/email or password")
	}

	accessToken, err := context.MakeToken(u)
	if err != nil {
		log.Errorf("Error creating token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, map[string]string{"token": "Bearer " + accessToken})
}

// Register
// @Tags Auth
// @Summary Register
// @ID register
// @Router /api/register [post]
// @Accept json
// @Param body body registerForm true "register json"
// @Produce json
// @Success 200
func (h *Handler) register(c echo.Context) error {
	registerForm, err := newRegisterForm(c)
	if err != nil {
		return err
	}

	u, _ := h.userRepo.FindOneByEmail(registerForm.Email)
	if u != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Email already registered")
	}

	user := &repository.User{
		ID:           bson.NewObjectID(),
		Email:        registerForm.Email,
		Password:     util.CryptPassword(registerForm.Password),
		Name:         registerForm.Name,
		Position:     _const.PositionOther,
		Avatar:       "",
		Phone:        "",
		CreatedAt:    time.Now(),
		IsDeleted:    false,
		TotalProject: 0,
		TotalTask:    0,
	}

	doc, err := h.userRepo.Insert(user)
	if err != nil {
		log.Errorf("Error inserting user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}
	return c.JSON(http.StatusOK, doc)
}

// ForgotPassword
// @Tags Auth
// @Summary ForgotPassword
// @ID forgot-password
// @Router /api/forgot-password [post]
// @Accept json
// @Param body body forgotPasswordForm true "forgot password json"
// @Produce json
// @Success 200
func (h *Handler) forgotPassword(c echo.Context) error {
	passwordForm, err := newForgotPasswordForm(c)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindOneByEmail(passwordForm.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "Email not found")
		}
		log.Errorf("Error finding user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	newPassword := util.RandomString(10)
	u.Password = util.CryptPassword(newPassword)

	_, err = h.userRepo.Update(u)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error, please try again")
	}

	go func(email string, template string) {
		err = mail.SendMail(nil, []string{email}, "New Password", template)
		if err != nil {
			log.Errorf("Error sending email: %v", err)
			return
		}
	}(u.Email, mail.ForgotTemplate(newPassword))
	return c.JSON(http.StatusOK, map[string]string{"message": "New password has been sent to your email"})
}
