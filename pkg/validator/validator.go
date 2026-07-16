package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}

}

func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}

func FormatValidationError(err error) map[string]string {
	fields := make(map[string]string)

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		for _, fe := range verrs {
			fields[strings.ToLower(fe.Field())] = message(fe)
		}
	}
	return fields
}

func message(field validator.FieldError) string {
	switch field.Tag() {
	case "required":
		return "required"
	case "min":
		return fmt.Sprintf("%s must be %s", field.Value(), field.Param())
	case "max":
		return fmt.Sprintf("%s must be %s", field.Value(), field.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of %s", field.Value(), field.Param())
	default:
		return "invalid value"
	}
}
