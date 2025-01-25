package me

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
	minPhoneLength    = 2
	maxPhoneLength    = 20
)

type errorDoc struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type updateMeForm struct {
	Name     string `form:"name" json:"name"`
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
	Position string `form:"position" json:"position"`
	Phone    string `form:"phone" json:"phone"`
}

func newUpdateMeForm(c echo.Context) (*updateMeForm, error) {
	form := new(updateMeForm)
	if err := c.Bind(form); err != nil {
		log.Errorf("Error binding form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.ToLower(strings.TrimSpace(form.Email))
	form.Position = strings.TrimSpace(form.Position)
	form.Phone = strings.TrimSpace(form.Phone)

	var validationErrors []errorDoc

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

	// Validate position
	if form.Position == "" {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "position",
			Message: "Position cannot be empty",
		})
	}

	// Validate phone
	if len(form.Phone) < minPhoneLength || len(form.Phone) > maxPhoneLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "phone",
			Message: "Phone must be between 2 and 20 characters",
		})
	}
	return form, nil
}
