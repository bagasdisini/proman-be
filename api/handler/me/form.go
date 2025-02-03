package me

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"net/http"
	_const "proman-backend/internal/pkg/const"
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
	Name             string `form:"name" json:"name"`
	Phone            string `form:"phone" json:"phone"`
	Email            string `form:"email" json:"email"`
	Position         string `form:"position" json:"position"`
	VerificationCode string `form:"verification_code" json:"verification_code"`
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
	if len(form.Name) != 0 && (len(form.Name) < minNameLength || len(form.Name) > maxNameLength) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "name",
			Message: "Name must be between 1 and 50 characters",
		})
	}

	// Validate email
	if len(form.Email) != 0 && (len(form.Email) < minEmailLength || len(form.Email) > maxEmailLength) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Email must be between 3 and 50 characters",
		})
	} else if len(form.Email) != 0 && !govalidator.IsEmail(form.Email) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "email",
			Message: "Invalid email format",
		})
	}

	// Validate position
	if len(form.Position) != 0 && !_const.IsValidPosition(form.Position) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "position",
			Message: "Invalid position",
		})
	}

	// Validate phone
	if len(form.Phone) != 0 && (len(form.Phone) < minPhoneLength || len(form.Phone) > maxPhoneLength) {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "phone",
			Message: "Phone must be between 2 and 20 characters",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}

type updateMePasswordForm struct {
	OldPassword      string `form:"old_password" json:"old_password"`
	NewPassword      string `form:"new_password" json:"new_password"`
	ConfirmPassword  string `form:"confirm_password" json:"confirm_password"`
	VerificationCode string `form:"verification_code" json:"verification_code"`
}

func newUpdateMePasswordForm(c echo.Context) (*updateMePasswordForm, error) {
	form := new(updateMePasswordForm)
	if err := c.Bind(form); err != nil {
		log.Errorf("Error binding form: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	form.OldPassword = strings.TrimSpace(form.OldPassword)
	form.NewPassword = strings.TrimSpace(form.NewPassword)
	form.ConfirmPassword = strings.TrimSpace(form.ConfirmPassword)
	form.VerificationCode = strings.TrimSpace(form.VerificationCode)

	var validationErrors []errorDoc

	// Validate old password
	if len(form.OldPassword) < minPasswordLength || len(form.OldPassword) > maxPasswordLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "old_password",
			Message: "Old password must be between 6 and 50 characters",
		})
	}

	// Validate new password
	if len(form.NewPassword) < minPasswordLength || len(form.NewPassword) > maxPasswordLength {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "new_password",
			Message: "New password must be between 6 and 50 characters",
		})
	}

	// Validate confirm password
	if form.NewPassword != form.ConfirmPassword {
		validationErrors = append(validationErrors, errorDoc{
			Field:   "confirm_password",
			Message: "Confirm password does not match",
		})
	}

	if len(validationErrors) > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}
	return form, nil
}
