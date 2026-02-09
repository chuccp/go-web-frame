package validator

import (
	"context"
	"regexp"
	"strings"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var mobileRegex = regexp.MustCompile(`^1\d{10}$`)

func validateMobile(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	phone = strings.TrimPrefix(phone, "+86")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	fa := mobileRegex.MatchString(phone)
	log.Info("validateMobile", zap.String("phone", phone), zap.Bool("match", fa))
	return fa
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
	log.Info("init validator")
	return nil
}

func (v *Validator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err != nil {

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, e := range validationErrs {
				return errors.Errorf("验证错误 - %s 格式错误", e.Value())
			}
		}

	}
	return err
}
