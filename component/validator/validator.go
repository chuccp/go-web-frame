package validator

import (
	"context"
	"regexp"
	"strings"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/go-playground/validator/v10"
)

var mobileRegex = regexp.MustCompile(`^1\d{10}$`)

func validateMobile(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	phone = strings.TrimPrefix(phone, "+86")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	fa := mobileRegex.MatchString(phone)
	return fa
}
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return false
	}
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	if !hasLower {
		return false
	}
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if !hasUpper {
		return false
	}
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return false
	}
	return true
}

type Validator struct {
	validate *validator.Validate
}

func (v *Validator) Init(ctx context.Context, config config2.IConfig) error {
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

func (v *Validator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err != nil {

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, e := range validationErrs {
				if e.Tag() == "password" {
					return errors.Errorf("密码必须包含大小写字母和数字，且长度不能小于8")
				}
				return errors.Errorf("验证错误 - %s 格式错误", e.Value())
			}
		}

	}
	return err
}
