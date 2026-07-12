package validator

import (
	"fmt"
	"regexp"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/go-playground/validator/v10"
)

// Validation error codes.
const (
	CodePasswordRequired = 2001
	CodeMobileInvalid    = 2002
	CodeFieldInvalid     = 2003
)

// Pre-defined validation errors. Use errors.Is/As to match.
var (
	ErrPasswordRequired = &ValidationError{Code: CodePasswordRequired, Message: "password must contain uppercase and lowercase letters and digits, with a minimum length of 8"}
	ErrMobileInvalid    = &ValidationError{Code: CodeMobileInvalid, Message: "invalid mobile phone number format"}
	ErrFieldInvalid     = &ValidationError{Code: CodeFieldInvalid, Message: "validation failed"}
)

// ValidationError is a structured error for validation failures.
// Supports errors.Is matching by error code.
type ValidationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%d] validation failed on field '%s': %s", e.Code, e.Field, e.Message)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Is enables errors.Is matching by error code.
// Returns true if target is a *ValidationError with the same Code.
func (e *ValidationError) Is(target error) bool {
	if t, ok := target.(*ValidationError); ok {
		return t.Code == e.Code
	}
	return false
}

// NewValidationError creates a new ValidationError with the given code and message.
func NewValidationError(code int, msg string) *ValidationError {
	return &ValidationError{Code: code, Message: msg}
}

// IsValidationError returns true if err is or wraps a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// ValidationErrorCode extracts the error code from a ValidationError.
// Returns 0 if err is not a ValidationError.
func ValidationErrorCode(err error) int {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve.Code
	}
	return 0
}

// ---- internal validators ----

var mobileRegex = regexp.MustCompile(`^1\d{10}$`)

func validateMobile(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	phone = strings.TrimPrefix(phone, "+86")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return mobileRegex.MatchString(phone)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return false
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return false
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return false
	}
	if !regexp.MustCompile(`\d`).MatchString(password) {
		return false
	}
	return true
}

// ---- Validator ----

// Validator provides struct field validation with custom phone and password rules.
type Validator struct {
	validate *validator.Validate
}

func (v *Validator) Init(ctx *core.Context) error {
	v.validate = validator.New()
	err := v.validate.RegisterValidation("mobile", validateMobile)
	if err != nil {
		return err
	}
	err = v.validate.RegisterValidation("password", validatePassword)
	if err != nil {
		return err
	}
	log.Info("init validator")
	return nil
}

// Validate validates the given struct against registered validation rules.
// Returns a *ValidationError on failure, which can be checked with errors.Is/As.
func (v *Validator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}

	for _, e := range validationErrs {
		switch e.Tag() {
		case "password":
			return ErrPasswordRequired
		case "mobile":
			return ErrMobileInvalid
		default:
			return &ValidationError{
				Code:    CodeFieldInvalid,
				Message: "invalid format",
				Field:   e.Field(),
			}
		}
	}
	return err
}
