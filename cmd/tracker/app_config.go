package tracker

import (
	"os"
	"time"
)

// DateFormatConfig holds the date format configuration
type DateFormatConfig struct {
	InputFormat   string // Format for parsing CLI input
	DisplayFormat string // Format for displaying dates to users
	DBFormat      string // Format for database storage (always ISO)
}

// AppConfig holds the application configuration
type AppConfig struct {
	DateFormat  DateFormatConfig
	DefaultUnit string // Default weight unit (kg or lbs)
}

// Default configurations
const (
	DefaultInputFormat   = "02-01-2006" // DD-MM-YYYY
	DefaultDisplayFormat = "02-01-2006" // DD-MM-YYYY
	DBFormat             = "2006-01-02" // YYYY-MM-DD (ISO standard)
	DefaultUnit          = "kg"         // Default weight unit
)

// Supported format mappings
var formatMappings = map[string]string{
	"dd-mm-yyyy": "02-01-2006",
	"mm-dd-yyyy": "01-02-2006",
	"yyyy-mm-dd": "2006-01-02",
	"dd/mm/yyyy": "02/01/2006",
	"mm/dd/yyyy": "01/02/2006",
	"yyyy/mm/dd": "2006/01/02",
}

// GetAppConfig returns the application configuration based on environment variables
func GetAppConfig() AppConfig {
	return GetAppConfigFromEnv(os.Getenv)
}

// GetAppConfigFromEnv returns the application configuration using a custom environment function
// This allows for dependency injection and better testability
func GetAppConfigFromEnv(getEnv func(string) string) AppConfig {
	dateConfig := DateFormatConfig{
		InputFormat:   DefaultInputFormat,
		DisplayFormat: DefaultDisplayFormat,
		DBFormat:      DBFormat, // Always ISO for database
	}

	// Check for input format override
	if inputFormat := getEnv("DATE_INPUT_FORMAT"); inputFormat != "" {
		if goFormat, exists := formatMappings[inputFormat]; exists {
			dateConfig.InputFormat = goFormat
		}
	}

	// Check for display format override
	if displayFormat := getEnv("DATE_DISPLAY_FORMAT"); displayFormat != "" {
		if goFormat, exists := formatMappings[displayFormat]; exists {
			dateConfig.DisplayFormat = goFormat
		}
	}

	// Get default unit configuration
	defaultUnit := DefaultUnit
	if unit := getEnv("DEFAULT_UNIT"); unit != "" {
		if unit == "kg" || unit == "lbs" {
			defaultUnit = unit
		}
	}

	return AppConfig{
		DateFormat:  dateConfig,
		DefaultUnit: defaultUnit,
	}
}

// GetDateFormatConfig returns the date format configuration (for backward compatibility)
func GetDateFormatConfig() DateFormatConfig {
	return GetAppConfig().DateFormat
}

// GetDateFormatConfigFromEnv returns the date format configuration using a custom environment function
func GetDateFormatConfigFromEnv(getEnv func(string) string) DateFormatConfig {
	return GetAppConfigFromEnv(getEnv).DateFormat
}

// ParseDate parses a date string using the configured input format
func ParseDate(dateStr string) (time.Time, error) {
	config := GetDateFormatConfig()
	return time.Parse(config.InputFormat, dateStr)
}

// ParseDateFromEnv parses a date string using a custom environment function
func ParseDateFromEnv(dateStr string, getEnv func(string) string) (time.Time, error) {
	config := GetDateFormatConfigFromEnv(getEnv)
	return time.Parse(config.InputFormat, dateStr)
}

// FormatDate formats a time.Time using the configured display format
func FormatDate(t time.Time) string {
	config := GetDateFormatConfig()
	return t.Format(config.DisplayFormat)
}

// FormatDateFromEnv formats a time.Time using a custom environment function
func FormatDateFromEnv(t time.Time, getEnv func(string) string) string {
	config := GetDateFormatConfigFromEnv(getEnv)
	return t.Format(config.DisplayFormat)
}

// FormatDateForDB formats a time.Time for database storage (always ISO)
func FormatDateForDB(t time.Time) string {
	return t.Format(DBFormat)
}

// GetInputFormatDescription returns a human-readable description of the input format
func GetInputFormatDescription() string {
	config := GetDateFormatConfig()

	// Find the human-readable format name
	for name, goFormat := range formatMappings {
		if goFormat == config.InputFormat {
			return name
		}
	}

	// Fallback to showing the Go format
	return config.InputFormat
}

// GetInputFormatDescriptionFromEnv returns a human-readable description using a custom environment function
func GetInputFormatDescriptionFromEnv(getEnv func(string) string) string {
	config := GetDateFormatConfigFromEnv(getEnv)

	// Find the human-readable format name
	for name, goFormat := range formatMappings {
		if goFormat == config.InputFormat {
			return name
		}
	}

	// Fallback to showing the Go format
	return config.InputFormat
}

// GetDefaultUnit returns the configured default weight unit
func GetDefaultUnit() string {
	return GetAppConfig().DefaultUnit
}

// GetDefaultUnitFromEnv returns the configured default weight unit using a custom environment function
func GetDefaultUnitFromEnv(getEnv func(string) string) string {
	return GetAppConfigFromEnv(getEnv).DefaultUnit
}
