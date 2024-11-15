package domain

import (
	"fmt"
	"strings"
)

const (
	ErrorCodeBadRequest            = 400
	ErrorCodeModuleNotFound        = 2001
	ErrorCodeConfigNotFound        = 2002
	ErrorCodeInvalidConfig         = 2003
	ErrorCodeConfigVersionConflict = 2004
	ErrorCodeSchemaNotFound        = 2005
)

type ConfigValidationError struct {
	Details map[string]string
}

func NewConfigValidationError(details map[string]string) ConfigValidationError {
	return ConfigValidationError{
		Details: details,
	}
}

func (e ConfigValidationError) Error() string {
	descriptions := make([]string, 0, len(e.Details))
	for field, err := range e.Details {
		descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
	}
	return strings.Join(descriptions, "; ")
}
