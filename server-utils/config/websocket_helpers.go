package config

// WebsocketConfigValidator provides validation helpers for websocket configurations
type WebsocketConfigValidator struct{}

// NewWebsocketConfigValidator creates a new websocket config validator
func NewWebsocketConfigValidator() *WebsocketConfigValidator {
	return &WebsocketConfigValidator{}
}

// BuildWebsocketValidation creates validation for websocket configuration
func (w *WebsocketConfigValidator) BuildWebsocketValidation() ValidatorFunc {
	return func(config interface{}) error {
		type websocketConfig interface {
			GetWebsocketReadBufferSize() int
			GetWebsocketWriteBufferSize() int
			GetWebsocketMaxMessageSize() int64
		}
		
		if cfg, ok := config.(websocketConfig); ok {
			if err := ValidatePositive(cfg.GetWebsocketReadBufferSize(), "websocket.read_buffer_size"); err != nil {
				return err
			}
			if err := ValidatePositive(cfg.GetWebsocketWriteBufferSize(), "websocket.write_buffer_size"); err != nil {
				return err
			}
			if err := ValidatePositiveInt64(cfg.GetWebsocketMaxMessageSize(), "websocket.max_message_size"); err != nil {
				return err
			}
		}
		
		return nil
	}
}

// BuildDeltaValidation creates validation for Delta Exchange configuration
func (w *WebsocketConfigValidator) BuildDeltaValidation() ValidatorFunc {
	return func(config interface{}) error {
		type deltaConfig interface {
			IsDeltaEnabled() bool
			GetDeltaURL() string
			GetDeltaChannels() []string
			GetDeltaProductIDs() []string
			GetDeltaReconnectMax() int
		}
		
		if cfg, ok := config.(deltaConfig); ok {
			if cfg.IsDeltaEnabled() {
				if err := ValidateNonEmpty(cfg.GetDeltaURL(), "delta.url"); err != nil {
					return err
				}
				if err := ValidateNonEmptySlice(cfg.GetDeltaChannels(), "delta.channels"); err != nil {
					return err
				}
				if err := ValidateNonEmptySlice(cfg.GetDeltaProductIDs(), "delta.product_ids"); err != nil {
					return err
				}
				if err := ValidateNonNegative(cfg.GetDeltaReconnectMax(), "delta.reconnect_max"); err != nil {
					return err
				}
			}
		}
		
		return nil
	}
}

// BuildSecurityValidation creates validation for security configuration
func (w *WebsocketConfigValidator) BuildSecurityValidation() ValidatorFunc {
	return func(config interface{}) error {
		type securityConfig interface {
			IsRateLimitEnabled() bool
			GetRateLimitRequests() int
			GetRateLimitDuration() int
		}
		
		if cfg, ok := config.(securityConfig); ok {
			if cfg.IsRateLimitEnabled() {
				if err := ValidatePositive(cfg.GetRateLimitRequests(), "security.rate_limit_requests"); err != nil {
					return err
				}
				if err := ValidatePositive(cfg.GetRateLimitDuration(), "security.rate_limit_duration"); err != nil {
					return err
				}
			}
		}
		
		return nil
	}
}

// CombineValidators combines multiple validator functions into one
func CombineValidators(validators ...ValidatorFunc) ValidatorFunc {
	return func(config interface{}) error {
		for _, validator := range validators {
			if err := validator(config); err != nil {
				return err
			}
		}
		return nil
	}
} 