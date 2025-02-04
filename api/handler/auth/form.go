package auth

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"net/http"
	"proman-backend/internal/pkg/log"
	"strings"
)

const (
	minNameLength     = 1
	maxNameLength     = 50
	minEmailLength    = 3
	maxEmailLength    = 50
	minPasswordLength = 6
	maxPasswordLength = 50
)

type errorDoc struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type loginForm struct {
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
}

func newLoginForm(c echo.Context) (*loginForm, error) {
	form := new(loginForm)
	if err := c.Bind(form); err != nil {
		log.Errorf("Error binding form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	form.Email = strings.ToLower(strings.TrimSpace(form.Email))

	validationErrors := make([]errorDoc, 0)

	// Validate email
	if len(form.Email) < minEmailLength || len(form.Email) > maxEmailLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Email must be between 3 and 50 characters",
		})
	} else if !govalidator.IsEmail(form.Email) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Invalid email format",
		})
	}

	// Validate password
	if len(form.Password) < minPasswordLength || len(form.Password) > maxPasswordLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "password",
			Message: "Password must be between 6 and 50 characters",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
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
		log.Errorf("Error binding form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.ToLower(strings.TrimSpace(form.Email))

	validationErrors := make([]errorDoc, 0)

	// Validate name
	if len(form.Name) < minNameLength || len(form.Name) > maxNameLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "name",
			Message: "Name must be between 1 and 50 characters",
		})
	}

	// Validate email
	if len(form.Email) < minEmailLength || len(form.Email) > maxEmailLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Email must be between 3 and 50 characters",
		})
	} else if !govalidator.IsEmail(form.Email) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Invalid email format",
		})
	}

	// Validate password
	if len(form.Password) < minPasswordLength || len(form.Password) > maxPasswordLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "password",
			Message: "Password must be between 6 and 50 characters",
		})
	}

	// Validate confirm password
	if form.Password != form.ConfirmPassword {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "confirm_password",
			Message: "Password and confirm password must be match",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}

type forgotPasswordForm struct {
	Email string `form:"email" json:"email"`
}

func newForgotPasswordForm(c echo.Context) (*forgotPasswordForm, error) {
	form := new(forgotPasswordForm)
	if err := c.Bind(form); err != nil {
		log.Errorf("Error binding form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	form.Email = strings.ToLower(strings.TrimSpace(form.Email))

	validationErrors := make([]errorDoc, 0)

	// Validate email
	if len(form.Email) < minEmailLength || len(form.Email) > maxEmailLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Email must be between 3 and 50 characters",
		})
	} else if !govalidator.IsEmail(form.Email) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Invalid email format",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}
