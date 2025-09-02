package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Service      ServiceConfig      `mapstructure:"service"`
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	RabbitMQ     RabbitMQConfig     `mapstructure:"rabbitmq"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	AWS          AWSConfig          `mapstructure:"aws"`
	I18n         I18nConfig         `mapstructure:"i18n"`
	OAuth        OAuthConfig        `mapstructure:"oauth"`
	Payment      PaymentConfig      `mapstructure:"payment"`
	Subscription SubscriptionConfig `mapstructure:"subscription"`
	Billing      BillingConfig      `mapstructure:"billing"`
}

type ServiceConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type RabbitMQConfig struct {
	URL            string        `mapstructure:"url"`
	Exchange       string        `mapstructure:"exchange"`
	Queue          string        `mapstructure:"queue"`
	ConsumerTag    string        `mapstructure:"consumer_tag"`
	PrefetchCount  int           `mapstructure:"prefetch_count"`
	ReconnectDelay time.Duration `mapstructure:"reconnect_delay"`
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessDuration  time.Duration `mapstructure:"access_duration"`
	RefreshDuration time.Duration `mapstructure:"refresh_duration"`
	Issuer          string        `mapstructure:"issuer"`
}

type AWSConfig struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	S3Bucket        string `mapstructure:"s3_bucket"`
}

type I18nConfig struct {
	DefaultLanguage    string   `mapstructure:"default_language"`
	SupportedLanguages []string `mapstructure:"supported_languages"`
}

type OAuthConfig struct {
	Google      OAuthProviderConfig `mapstructure:"google"`
	Facebook    OAuthProviderConfig `mapstructure:"facebook"`
	StateExpiry time.Duration       `mapstructure:"state_expiry"`
	PKCEEnabled bool                `mapstructure:"pkce_enabled"`
}

type OAuthProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	AppID        string   `mapstructure:"app_id"`     // For Facebook
	AppSecret    string   `mapstructure:"app_secret"` // For Facebook
	RedirectURI  string   `mapstructure:"redirect_uri"`
	Scopes       []string `mapstructure:"scopes"`
}

type PaymentConfig struct {
	Provider string       `mapstructure:"provider"` // stripe, paypal, mock
	Stripe   StripeConfig `mapstructure:"stripe"`
}

type StripeConfig struct {
	SecretKey      string `mapstructure:"secret_key"`
	PublishableKey string `mapstructure:"publishable_key"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
}

type SubscriptionConfig struct {
	TrialDaysDefault        int    `mapstructure:"trial_days_default"`
	GracePeriodDays         int    `mapstructure:"grace_period_days"`
	WebhookSecret           string `mapstructure:"webhook_secret"`
	EnableProrations        bool   `mapstructure:"enable_prorations"`
	AllowImmediateDowngrade bool   `mapstructure:"allow_immediate_downgrades"`
}

type BillingConfig struct {
	ProcessHour              int   `mapstructure:"process_hour"`
	RetryAttempts            int   `mapstructure:"retry_attempts"`
	RetryDelayHours          int   `mapstructure:"retry_delay_hours"`
	ReminderDaysBeforeExpiry []int `mapstructure:"reminder_days_before_expiry"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("MYTORRA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("service.environment", "development")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", "5m")

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	v.SetDefault("rabbitmq.prefetch_count", 1)
	v.SetDefault("rabbitmq.reconnect_delay", "5s")

	v.SetDefault("jwt.secret", "change-this-secret-in-production")
	v.SetDefault("jwt.access_duration", "15m")
	v.SetDefault("jwt.refresh_duration", "7d")
	v.SetDefault("jwt.issuer", "mytorra")

	v.SetDefault("i18n.default_language", "en")
	v.SetDefault("i18n.supported_languages", []string{"en", "fr"})

	v.SetDefault("oauth.state_expiry", "10m")
	v.SetDefault("oauth.pkce_enabled", true)

	v.SetDefault("payment.provider", "mock")

	v.SetDefault("subscription.trial_days_default", 7)
	v.SetDefault("subscription.grace_period_days", 3)
	v.SetDefault("subscription.enable_prorations", true)
	v.SetDefault("subscription.allow_immediate_downgrades", false)

	v.SetDefault("billing.process_hour", 2)
	v.SetDefault("billing.retry_attempts", 3)
	v.SetDefault("billing.retry_delay_hours", 24)
	v.SetDefault("billing.reminder_days_before_expiry", []int{7, 3, 1})
}

func validate(cfg *Config) error {
	if cfg.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if cfg.Database.Host != "" && cfg.Database.Database == "" {
		return fmt.Errorf("database name is required when host is specified")
	}

	return nil
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode)
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Environment variable utility functions

// GetEnvOrDefault returns the value of the environment variable or the default value if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvOrDefaultInt returns the value of the environment variable as an integer or the default value if not set or invalid
func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvOrDefaultBool returns the value of the environment variable as a boolean or the default value if not set or invalid
func GetEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
