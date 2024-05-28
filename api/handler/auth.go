package handler

import (
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"proman-backend/api/repository"
	_const "proman-backend/pkg/const"
	"proman-backend/pkg/context"
	"proman-backend/pkg/mail"
	"proman-backend/pkg/util"
	"strings"
	"time"
)

type loginForm struct {
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
}

func newLoginForm(c echo.Context) (*loginForm, error) {
	form := new(loginForm)
	if err := c.Bind(form); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid email or password")
	}

	form.Email = strings.ToLower(strings.TrimSpace(form.Email))

	if len(form.Email) < 3 || len(form.Email) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Email must be between 3 and 50 characters")
	}
	if !govalidator.IsEmail(form.Email) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid email")
	}

	if len(form.Password) < 6 || len(form.Password) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Password must be between 6 and 50 characters")
	}
	return form, nil
}

type registerForm struct {
	Name            string `form:"name" json:"name"`
	Email           string `form:"email" json:"email"`
	Password        string `form:"password" json:"password"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password"`
}

func newRegisterForm(c echo.Context) (*registerForm, error) {
	form := new(registerForm)
	if err := c.Bind(form); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid email or password")
	}

	form.Name = strings.TrimSpace(form.Name)
	if len(form.Name) < 1 || len(form.Name) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Name must be between 1 and 50 characters")
	}

	form.Email = strings.ToLower(strings.TrimSpace(form.Email))

	if len(form.Email) < 3 || len(form.Email) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Email must be between 3 and 50 characters")
	}
	if !govalidator.IsEmail(form.Email) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid email")
	}

	if len(form.Password) < 6 || len(form.Password) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Password must be between 6 and 50 characters")
	}

	if form.Password != form.ConfirmPassword {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Password and confirm password must be the same")
	}
	return form, nil
}

type forgotPasswordForm struct {
	Email string `form:"email" json:"email"`
}

func newForgotPasswordForm(c echo.Context) (*forgotPasswordForm, error) {
	form := new(forgotPasswordForm)
	if err := c.Bind(form); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid email")
	}

	form.Email = strings.ToLower(strings.TrimSpace(form.Email))

	if len(form.Email) < 3 || len(form.Email) > 50 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Email must be between 3 and 50 characters")
	}
	if !govalidator.IsEmail(form.Email) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid email")
	}
	return form, nil
}

type AuthHandler struct {
	userRepo *repository.UserCollRepository
}

func NewAuthHandler(e *echo.Echo, db *mongo.Database) *AuthHandler {
	h := &AuthHandler{
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
// @Produce  json
// @Success 200
func (h *AuthHandler) login(c echo.Context) error {
	form, err := newLoginForm(c)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindOneByEmail(form.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Wrong username/email or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error when finding user")
	}

	if !util.CheckPassword(u.Password, form.Password) {
		return echo.NewHTTPError(http.StatusBadRequest, "Wrong username/email or password")
	}

	accessToken, err := context.MakeToken(u)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error when creating token")
	}
	return c.JSON(http.StatusOK, map[string]string{"token": accessToken})
}

// Register
// @Tags Auth
// @Summary Register
// @ID register
// @Router /api/register [post]
// @Accept json
// @Param body body registerForm true "register json"
// @Produce  json
// @Success 200
func (h *AuthHandler) register(c echo.Context) error {
	form, err := newRegisterForm(c)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindOneByEmail(form.Email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error when finding user")
	}

	if u != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Email already registered")
	}

	user := &repository.User{
		ID:        primitive.NewObjectID(),
		Email:     form.Email,
		Password:  util.CryptPassword(form.Password),
		Name:      form.Name,
		Role:      _const.RoleDeveloper,
		Position:  _const.PositionOther,
		CreatedAt: time.Now(),
		IsDeleted: false,
	}

	doc, err := h.userRepo.Insert(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error when inserting user")
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
// @Produce  json
// @Success 200
func (h *AuthHandler) forgotPassword(c echo.Context) error {
	form, err := newForgotPasswordForm(c)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindOneByEmail(form.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "Email not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error when finding user")
	}

	newPassword := util.RandomString(10)
	u.Password = util.CryptPassword(newPassword)

	_, err = h.userRepo.Update(u)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "There was an error when updating user")
	}

	go func(email string, template string) {
		err = mail.SendMail(nil, []string{email}, "New Password", template)
		if err != nil {
			return
		}
	}(u.Email, mail.ForgotTemplate(newPassword))

	return c.JSON(http.StatusOK, map[string]string{"message": "New password has been sent to your email"})
}
