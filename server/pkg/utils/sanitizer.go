package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

var htmlRegex = regexp.MustCompile(`(?s)<.*?>`)

var NoHTMLValidator validator.Func = func(fl validator.FieldLevel) bool {
	input := fl.Field().String()

	return !htmlRegex.MatchString(input)
}

var strictPolicy = bluemonday.StrictPolicy()

func SanitizeString(input string) string {
	return strictPolicy.Sanitize(input)
}
