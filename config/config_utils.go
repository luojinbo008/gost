package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func verify(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		errs := err.(validator.ValidationErrors)
		var slice []string
		for _, msg := range errs {
			slice = append(slice, msg.Error())
		}
		return errors.New(strings.Join(slice, ","))
	}
	return nil
}

func translateIds(registryIds []string) []string {
	ids := make([]string, 0)
	for _, id := range registryIds {

		ids = append(ids, strings.Split(id, ",")...)
	}
	return removeDuplicateElement(ids)
}

// removeDuplicateElement remove duplicate element
func removeDuplicateElement(items []string) []string {
	result := make([]string, 0, len(items))
	temp := map[string]struct{}{}
	for _, item := range items {
		if _, ok := temp[item]; !ok && item != "" {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// clientNameID unique identifier id for client
func clientNameID(protocol, address string) string {
	return strings.Join([]string{protocol, address}, "-")
}
