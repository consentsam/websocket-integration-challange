package config

import (
	"fmt"
)

// ValidatorFunc is a function that validates a specific configuration aspect
type ValidatorFunc func(config interface{}) error

// ConfigValidator provides validation for configuration structs
type ConfigValidator struct {
	validators []ValidatorFunc
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		validators: make([]ValidatorFunc, 0),
	}
}

// AddValidator adds a validation function
func (cv *ConfigValidator) AddValidator(validator ValidatorFunc) {
	cv.validators = append(cv.validators, validator)
}

// Validate runs all validators on the configuration
func (cv *ConfigValidator) Validate(config interface{}) error {
	for _, validator := range cv.validators {
		if err := validator(config); err != nil {
			return err
		}
	}
	return nil
}

// Common validation functions

// ValidatePortRange validates that a port is within valid range
func ValidatePortRange(port int, name string) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535, got %d", name, port)
	}
	return nil
}

// ValidatePositive validates that a number is positive
func ValidatePositive(value int, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %d", name, value)
	}
	return nil
}

// ValidatePositiveInt64 validates that an int64 is positive
func ValidatePositiveInt64(value int64, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %d", name, value)
	}
	return nil
}

// ValidateNonEmpty validates that a string is not empty
func ValidateNonEmpty(value string, name string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ValidateNonEmptySlice validates that a slice is not empty
func ValidateNonEmptySlice(value []string, name string) error {
	if len(value) == 0 {
		return fmt.Errorf("%s must not be empty", name)
	}
	return nil
}

// ValidateNonNegative validates that a number is non-negative
func ValidateNonNegative(value int, name string) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative, got %d", name, value)
	}
	return nil
} 