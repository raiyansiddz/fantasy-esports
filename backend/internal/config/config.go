package config

import (
	"fantasy-esports-backend/config"
)

// Config wraps the main config for internal services
type Config struct {
	*config.Config
}

// New creates a new internal config
func New() *Config {
	return &Config{
		Config: config.Load(),
	}
}

// GetDatabaseURL returns the database URL
func (c *Config) GetDatabaseURL() string {
	return c.DatabaseURL
}

// GetJWTSecret returns the JWT secret
func (c *Config) GetJWTSecret() string {
	return c.JWTSecret
}

// GetPort returns the server port
func (c *Config) GetPort() string {
	return c.Port
}

// GetCloudinaryURL returns the Cloudinary URL
func (c *Config) GetCloudinaryURL() string {
	return c.CloudinaryURL
}

// GetGinMode returns the Gin mode
func (c *Config) GetGinMode() string {
	return c.GinMode
}