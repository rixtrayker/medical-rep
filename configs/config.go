package configs

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	// "github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `koanf:"app"`
	HTTP     HTTPConfig     `koanf:"http"`
	Database DatabaseConfig `koanf:"database"`
	Redis    RedisConfig    `koanf:"redis"`
	Auth     AuthConfig     `koanf:"auth"`
	Logging  LoggingConfig  `koanf:"logging"`
	Health   HealthConfig   `koanf:"health"`
}

type AppConfig struct {
	Name        string        `koanf:"name"`
	Version     string        `koanf:"version"`
	Environment string        `koanf:"environment"`
	Debug       bool          `koanf:"debug"`
	Shutdown    ShutdownConfig `koanf:"shutdown"`
}

type ShutdownConfig struct {
	Timeout time.Duration `koanf:"timeout"`
}

type HTTPConfig struct {
	Port            int           `koanf:"port"`
	Host            string        `koanf:"host"`
	ReadTimeout     time.Duration `koanf:"read_timeout"`
	WriteTimeout    time.Duration `koanf:"write_timeout"`
	IdleTimeout     time.Duration `koanf:"idle_timeout"`
	MaxHeaderBytes  int           `koanf:"max_header_bytes"`
	TLS             TLSConfig     `koanf:"tls"`
	CORS            CORSConfig    `koanf:"cors"`
	RateLimit       RateLimitConfig `koanf:"rate_limit"`
}

type TLSConfig struct {
	Enabled  bool   `koanf:"enabled"`
	CertFile string `koanf:"cert_file"`
	KeyFile  string `koanf:"key_file"`
}

type CORSConfig struct {
	AllowedOrigins []string `koanf:"allowed_origins"`
	AllowedMethods []string `koanf:"allowed_methods"`
	AllowedHeaders []string `koanf:"allowed_headers"`
}

type RateLimitConfig struct {
	Enabled bool    `koanf:"enabled"`
	Rate    float64 `koanf:"rate"`
	Burst   int     `koanf:"burst"`
}

type DatabaseConfig struct {
	Driver          string        `koanf:"driver"`
	Host            string        `koanf:"host"`
	Port            int           `koanf:"port"`
	Database        string        `koanf:"database"`
	Username        string        `koanf:"username"`
	Password        string        `koanf:"password"`
	SSLMode         string        `koanf:"ssl_mode"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
	MigrationsPath  string        `koanf:"migrations_path"`
}

type RedisConfig struct {
	Host         string        `koanf:"host"`
	Port         int           `koanf:"port"`
	Password     string        `koanf:"password"`
	Database     int           `koanf:"database"`
	PoolSize     int           `koanf:"pool_size"`
	DialTimeout  time.Duration `koanf:"dial_timeout"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

type AuthConfig struct {
	JWTSecret     string        `koanf:"jwt_secret"`
	JWTExpiration time.Duration `koanf:"jwt_expiration"`
	BCryptCost    int           `koanf:"bcrypt_cost"`
}

type LoggingConfig struct {
	Level      string `koanf:"level"`
	Format     string `koanf:"format"`
	Output     string `koanf:"output"`
	MaxSize    int    `koanf:"max_size"`
	MaxBackups int    `koanf:"max_backups"`
	MaxAge     int    `koanf:"max_age"`
	Compress   bool   `koanf:"compress"`
}

type HealthConfig struct {
	Enabled         bool          `koanf:"enabled"`
	CheckInterval   time.Duration `koanf:"check_interval"`
	Timeout         time.Duration `koanf:"timeout"`
	DatabaseCheck   bool          `koanf:"database_check"`
	RedisCheck      bool          `koanf:"redis_check"`
	ExternalChecks  []string      `koanf:"external_checks"`
}

var (
	k *koanf.Koanf
	C *Config
)

// Load initializes and loads configuration from multiple sources
func Load() error {
	k = koanf.New(".")
	
	// 1. Load default values
	if err := loadDefaults(); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	// 2. Load base configuration file
	if err := loadConfigFile("configs/config.yaml"); err != nil {
		log.Printf("Warning: Could not load base config file: %v", err)
	}

	// 3. Load environment-specific configuration
	env := k.String("app.environment")
	envConfigFile := fmt.Sprintf("configs/config.%s.yaml", env)
	if err := loadConfigFile(envConfigFile); err != nil {
		log.Printf("Warning: Could not load environment config file %s: %v", envConfigFile, err)
	}

	// 4. Load environment variables
	if err := loadEnvVars(); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// 5. Unmarshal into config struct
	C = &Config{}
	if err := k.Unmarshal("", C); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. Validate configuration
	if err := validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

func loadDefaults() error {
	defaults := Config{
		App: AppConfig{
			Name:        "medical-rep-api",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       true,
			Shutdown: ShutdownConfig{
				Timeout: 30 * time.Second,
			},
		},
		HTTP: HTTPConfig{
			Port:           8080,
			Host:           "0.0.0.0",
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1MB
			TLS: TLSConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"*"},
			},
			RateLimit: RateLimitConfig{
				Enabled: false,
				Rate:    100,
				Burst:   200,
			},
		},
		Database: DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Database:        "medical_rep",
			Username:        "postgres",
			Password:        "password",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			MigrationsPath:  "migrations",
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Database:     0,
			PoolSize:     10,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Auth: AuthConfig{
			JWTExpiration: 24 * time.Hour,
			BCryptCost:    12,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		Health: HealthConfig{
			Enabled:        true,
			CheckInterval:  30 * time.Second,
			Timeout:        5 * time.Second,
			DatabaseCheck:  true,
			RedisCheck:     true,
			ExternalChecks: []string{},
		},
	}

	return k.Load(structs.Provider(defaults, "koanf"), nil)
}

func loadConfigFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return k.Load(file.Provider(path), yaml.Parser())
}

func loadEnvVars() error {
	return k.Load(env.Provider("", ".", func(s string) string {
		// Convert MEDICAL_REP_APP_NAME to app.name
		s = strings.TrimPrefix(s, "MEDICAL_REP_")
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	}), nil)
}

func validate() error {
	// Validate required fields
	if C.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if C.HTTP.Port <= 0 || C.HTTP.Port > 65535 {
		return fmt.Errorf("http.port must be between 1 and 65535")
	}

	if C.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}

	if C.Auth.JWTSecret == "" && C.App.Environment == "production" {
		return fmt.Errorf("auth.jwt_secret is required in production")
	}

	// Validate TLS configuration
	if C.HTTP.TLS.Enabled {
		if C.HTTP.TLS.CertFile == "" || C.HTTP.TLS.KeyFile == "" {
			return fmt.Errorf("tls.cert_file and tls.key_file are required when TLS is enabled")
		}
	}

	return nil
}

// Get returns the global configuration instance
func Get() *Config {
	if C == nil {
		log.Fatal("Configuration not loaded. Call config.Load() first.")
	}
	return C
}

// GetConnectionString returns the database connection string
func (c *Config) GetConnectionString() string {
	switch c.Database.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Database,
			c.Database.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
		)
	default:
		return ""
	}
}

// GetRedisAddr returns Redis address in host:port format
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}