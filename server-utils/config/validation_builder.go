package config

import "fmt"

// ValidationBuilder helps build validation logic for configuration structs
type ValidationBuilder struct {
	validators []ValidatorFunc
}

// NewValidationBuilder creates a new validation builder
func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{
		validators: make([]ValidatorFunc, 0),
	}
}

// AddCustom adds a custom validation function
func (vb *ValidationBuilder) AddCustom(validator ValidatorFunc) *ValidationBuilder {
	vb.validators = append(vb.validators, validator)
	return vb
}

// AddFieldValidator adds a validation for a specific field
func (vb *ValidationBuilder) AddFieldValidator(fieldGetter func(interface{}) interface{}, validator func(interface{}) error) *ValidationBuilder {
	vb.validators = append(vb.validators, func(config interface{}) error {
		value := fieldGetter(config)
		return validator(value)
	})
	return vb
}

// RequireNonEmpty adds a non-empty string validation
func (vb *ValidationBuilder) RequireNonEmpty(fieldGetter func(interface{}) string, fieldName string) *ValidationBuilder {
	return vb.AddFieldValidator(
		func(c interface{}) interface{} { return fieldGetter(c) },
		func(v interface{}) error { return ValidateNonEmpty(v.(string), fieldName) },
	)
}

// RequirePort adds a port validation
func (vb *ValidationBuilder) RequirePort(fieldGetter func(interface{}) int, fieldName string) *ValidationBuilder {
	return vb.AddFieldValidator(
		func(c interface{}) interface{} { return fieldGetter(c) },
		func(v interface{}) error { return ValidatePortRange(v.(int), fieldName) },
	)
}

// RequirePositive adds a positive number validation
func (vb *ValidationBuilder) RequirePositive(fieldGetter func(interface{}) int, fieldName string) *ValidationBuilder {
	return vb.AddFieldValidator(
		func(c interface{}) interface{} { return fieldGetter(c) },
		func(v interface{}) error { return ValidatePositive(v.(int), fieldName) },
	)
}

// RequirePositiveInt64 adds a positive int64 validation
func (vb *ValidationBuilder) RequirePositiveInt64(fieldGetter func(interface{}) int64, fieldName string) *ValidationBuilder {
	return vb.AddFieldValidator(
		func(c interface{}) interface{} { return fieldGetter(c) },
		func(v interface{}) error { return ValidatePositiveInt64(v.(int64), fieldName) },
	)
}

// RequireNonNegative adds a non-negative number validation
func (vb *ValidationBuilder) RequireNonNegative(fieldGetter func(interface{}) int, fieldName string) *ValidationBuilder {
	return vb.AddFieldValidator(
		func(c interface{}) interface{} { return fieldGetter(c) },
		func(v interface{}) error { return ValidateNonNegative(v.(int), fieldName) },
	)
}

// RequireNonEmptySlice adds a non-empty slice validation
func (vb *ValidationBuilder) RequireNonEmptySlice(fieldGetter func(interface{}) []string, fieldName string) *ValidationBuilder {
	return vb.AddFieldValidator(
		func(c interface{}) interface{} { return fieldGetter(c) },
		func(v interface{}) error { return ValidateNonEmptySlice(v.([]string), fieldName) },
	)
}

// RequireConditional adds a conditional validation
func (vb *ValidationBuilder) RequireConditional(condition func(interface{}) bool, validator ValidatorFunc) *ValidationBuilder {
	vb.validators = append(vb.validators, func(config interface{}) error {
		if condition(config) {
			return validator(config)
		}
		return nil
	})
	return vb
}

// Build creates a validator function from all the added validations
func (vb *ValidationBuilder) Build() ValidatorFunc {
	return func(config interface{}) error {
		for _, validator := range vb.validators {
			if err := validator(config); err != nil {
				return err
			}
		}
		return nil
	}
}

// Common validation builders for typical configurations

// BuildServiceValidation creates standard service configuration validation
func BuildServiceValidation() ValidatorFunc {
	return NewValidationBuilder().
		RequireNonEmpty(func(c interface{}) string {
			if cfg, ok := c.(interface{ GetServiceName() string }); ok {
				return cfg.GetServiceName()
			}
			return ""
		}, "service_name").
		RequirePort(func(c interface{}) int {
			if cfg, ok := c.(interface{ GetHTTPPort() int }); ok {
				return cfg.GetHTTPPort()
			}
			return 0
		}, "http_port").
		RequirePort(func(c interface{}) int {
			if cfg, ok := c.(interface{ GetGRPCPort() int }); ok {
				return cfg.GetGRPCPort()
			}
			return 0
		}, "grpc_port").
		AddCustom(func(c interface{}) error {
			type portsGetter interface {
				GetHTTPPort() int
				GetGRPCPort() int
			}
			if cfg, ok := c.(portsGetter); ok {
				if cfg.GetHTTPPort() == cfg.GetGRPCPort() {
					return fmt.Errorf("http_port and grpc_port cannot be the same (%d)", cfg.GetHTTPPort())
				}
			}
			return nil
		}).
		Build()
} 