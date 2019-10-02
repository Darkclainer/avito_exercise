package main

import (
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

func NewValidate() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("identificator", validateRegexp(`^[a-zA-Z]\w*$`))
	validate.RegisterAlias("username", "identificator,min=1,max=32")
	validate.RegisterAlias("chatname", "identificator,min=1,max=32")
	return validate
}

func validateRegexp(regexpRaw string) validator.Func {
	template := regexp.MustCompile(regexpRaw)
	return func(fl validator.FieldLevel) bool {
		return template.MatchString(fl.Field().String())
	}
}
