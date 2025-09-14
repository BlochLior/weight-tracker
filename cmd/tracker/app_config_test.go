package tracker

import (
	"testing"
	"time"
)

func TestGetAppConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig AppConfig
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expectedConfig: AppConfig{
				DateFormat: DateFormatConfig{
					InputFormat:   DefaultInputFormat,
					DisplayFormat: DefaultDisplayFormat,
					DBFormat:      DBFormat,
				},
				DefaultUnit: DefaultUnit,
			},
		},
		{
			name: "custom date formats",
			envVars: map[string]string{
				"DATE_INPUT_FORMAT":   "mm-dd-yyyy",
				"DATE_DISPLAY_FORMAT": "yyyy-mm-dd",
			},
			expectedConfig: AppConfig{
				DateFormat: DateFormatConfig{
					InputFormat:   "01-02-2006",
					DisplayFormat: "2006-01-02",
					DBFormat:      DBFormat,
				},
				DefaultUnit: DefaultUnit,
			},
		},
		{
			name: "custom unit",
			envVars: map[string]string{
				"DEFAULT_UNIT": "lbs",
			},
			expectedConfig: AppConfig{
				DateFormat: DateFormatConfig{
					InputFormat:   DefaultInputFormat,
					DisplayFormat: DefaultDisplayFormat,
					DBFormat:      DBFormat,
				},
				DefaultUnit: "lbs",
			},
		},
		{
			name: "all custom settings",
			envVars: map[string]string{
				"DATE_INPUT_FORMAT":   "yyyy-mm-dd",
				"DATE_DISPLAY_FORMAT": "dd/mm/yyyy",
				"DEFAULT_UNIT":        "lbs",
			},
			expectedConfig: AppConfig{
				DateFormat: DateFormatConfig{
					InputFormat:   "2006-01-02",
					DisplayFormat: "02/01/2006",
					DBFormat:      DBFormat,
				},
				DefaultUnit: "lbs",
			},
		},
		{
			name: "invalid unit falls back to default",
			envVars: map[string]string{
				"DEFAULT_UNIT": "invalid",
			},
			expectedConfig: AppConfig{
				DateFormat: DateFormatConfig{
					InputFormat:   DefaultInputFormat,
					DisplayFormat: DefaultDisplayFormat,
					DBFormat:      DBFormat,
				},
				DefaultUnit: DefaultUnit, // Should fall back to default
			},
		},
		{
			name: "invalid date format falls back to default",
			envVars: map[string]string{
				"DATE_INPUT_FORMAT": "invalid-format",
			},
			expectedConfig: AppConfig{
				DateFormat: DateFormatConfig{
					InputFormat:   DefaultInputFormat, // Should fall back to default
					DisplayFormat: DefaultDisplayFormat,
					DBFormat:      DBFormat,
				},
				DefaultUnit: DefaultUnit,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock environment function
			getEnv := func(key string) string {
				return tt.envVars[key]
			}

			// Get configuration using dependency injection
			config := GetAppConfigFromEnv(getEnv)

			// Verify configuration
			if config.DateFormat.InputFormat != tt.expectedConfig.DateFormat.InputFormat {
				t.Errorf("InputFormat = %v, want %v", config.DateFormat.InputFormat, tt.expectedConfig.DateFormat.InputFormat)
			}
			if config.DateFormat.DisplayFormat != tt.expectedConfig.DateFormat.DisplayFormat {
				t.Errorf("DisplayFormat = %v, want %v", config.DateFormat.DisplayFormat, tt.expectedConfig.DateFormat.DisplayFormat)
			}
			if config.DateFormat.DBFormat != tt.expectedConfig.DateFormat.DBFormat {
				t.Errorf("DBFormat = %v, want %v", config.DateFormat.DBFormat, tt.expectedConfig.DateFormat.DBFormat)
			}
			if config.DefaultUnit != tt.expectedConfig.DefaultUnit {
				t.Errorf("DefaultUnit = %v, want %v", config.DefaultUnit, tt.expectedConfig.DefaultUnit)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name         string
		dateStr      string
		envVar       string
		expectError  bool
		expectedDay  int
		expectedMon  int
		expectedYear int
	}{
		{
			name:         "default format dd-mm-yyyy",
			dateStr:      "15-09-2024",
			envVar:       "",
			expectError:  false,
			expectedDay:  15,
			expectedMon:  9,
			expectedYear: 2024,
		},
		{
			name:         "mm-dd-yyyy format",
			dateStr:      "09-15-2024",
			envVar:       "mm-dd-yyyy",
			expectError:  false,
			expectedDay:  15,
			expectedMon:  9,
			expectedYear: 2024,
		},
		{
			name:         "yyyy-mm-dd format",
			dateStr:      "2024-09-15",
			envVar:       "yyyy-mm-dd",
			expectError:  false,
			expectedDay:  15,
			expectedMon:  9,
			expectedYear: 2024,
		},
		{
			name:         "dd/mm/yyyy format",
			dateStr:      "15/09/2024",
			envVar:       "dd/mm/yyyy",
			expectError:  false,
			expectedDay:  15,
			expectedMon:  9,
			expectedYear: 2024,
		},
		{
			name:        "invalid date format",
			dateStr:     "15-09-2024",
			envVar:      "mm-dd-yyyy", // Wrong format for the date string
			expectError: true,
		},
		{
			name:        "invalid date string",
			dateStr:     "invalid-date",
			envVar:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock environment function
			getEnv := func(key string) string {
				if key == "DATE_INPUT_FORMAT" {
					return tt.envVar
				}
				return ""
			}

			// Parse date using dependency injection
			result, err := ParseDateFromEnv(tt.dateStr, getEnv)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseDateFromEnv() expected error but got none")
				}
			} else {
				if err != nil {
					t.Error(unexpectedErrorString(err))
					return
				}

				// Check parsed date components
				if result.Day() != tt.expectedDay {
					t.Errorf("Day = %v, want %v", result.Day(), tt.expectedDay)
				}
				if result.Month() != time.Month(tt.expectedMon) {
					t.Errorf("Month = %v, want %v", result.Month(), tt.expectedMon)
				}
				if result.Year() != tt.expectedYear {
					t.Errorf("Year = %v, want %v", result.Year(), tt.expectedYear)
				}
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	testDate := time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		envVar         string
		expectedFormat string
	}{
		{
			name:           "default format dd-mm-yyyy",
			envVar:         "",
			expectedFormat: "15-09-2024",
		},
		{
			name:           "mm-dd-yyyy format",
			envVar:         "mm-dd-yyyy",
			expectedFormat: "09-15-2024",
		},
		{
			name:           "yyyy-mm-dd format",
			envVar:         "yyyy-mm-dd",
			expectedFormat: "2024-09-15",
		},
		{
			name:           "dd/mm/yyyy format",
			envVar:         "dd/mm/yyyy",
			expectedFormat: "15/09/2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock environment function
			getEnv := func(key string) string {
				if key == "DATE_DISPLAY_FORMAT" {
					return tt.envVar
				}
				return ""
			}

			// Format date using dependency injection
			result := FormatDateFromEnv(testDate, getEnv)

			// Check result
			if result != tt.expectedFormat {
				t.Errorf("FormatDateFromEnv() = %v, want %v", result, tt.expectedFormat)
			}
		})
	}
}

func TestFormatDateForDB(t *testing.T) {
	testDate := time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC)
	expected := "2024-09-15"

	result := FormatDateForDB(testDate)
	if result != expected {
		t.Errorf("FormatDateForDB() = %v, want %v", result, expected)
	}
}

func TestGetDefaultUnit(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		expectedUnit string
	}{
		{
			name:         "default unit kg",
			envVar:       "",
			expectedUnit: "kg",
		},
		{
			name:         "custom unit lbs",
			envVar:       "lbs",
			expectedUnit: "lbs",
		},
		{
			name:         "invalid unit falls back to default",
			envVar:       "invalid",
			expectedUnit: "kg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock environment function
			getEnv := func(key string) string {
				if key == "DEFAULT_UNIT" {
					return tt.envVar
				}
				return ""
			}

			// Get default unit using dependency injection
			result := GetDefaultUnitFromEnv(getEnv)

			// Check result
			if result != tt.expectedUnit {
				t.Errorf("GetDefaultUnitFromEnv() = %v, want %v", result, tt.expectedUnit)
			}
		})
	}
}

func TestGetInputFormatDescription(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		expectedDesc string
	}{
		{
			name:         "default format description",
			envVar:       "",
			expectedDesc: "dd-mm-yyyy",
		},
		{
			name:         "mm-dd-yyyy format description",
			envVar:       "mm-dd-yyyy",
			expectedDesc: "mm-dd-yyyy",
		},
		{
			name:         "yyyy-mm-dd format description",
			envVar:       "yyyy-mm-dd",
			expectedDesc: "yyyy-mm-dd",
		},
		{
			name:         "dd/mm/yyyy format description",
			envVar:       "dd/mm/yyyy",
			expectedDesc: "dd/mm/yyyy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock environment function
			getEnv := func(key string) string {
				if key == "DATE_INPUT_FORMAT" {
					return tt.envVar
				}
				return ""
			}

			// Get format description using dependency injection
			result := GetInputFormatDescriptionFromEnv(getEnv)

			// Check result
			if result != tt.expectedDesc {
				t.Errorf("GetInputFormatDescriptionFromEnv() = %v, want %v", result, tt.expectedDesc)
			}
		})
	}
}

func TestGetDateFormatConfig(t *testing.T) {
	// Test backward compatibility
	config := GetDateFormatConfig()

	// Should return the same as GetAppConfig().DateFormat
	appConfig := GetAppConfig()

	if config.InputFormat != appConfig.DateFormat.InputFormat {
		t.Errorf("GetDateFormatConfig().InputFormat = %v, want %v", config.InputFormat, appConfig.DateFormat.InputFormat)
	}
	if config.DisplayFormat != appConfig.DateFormat.DisplayFormat {
		t.Errorf("GetDateFormatConfig().DisplayFormat = %v, want %v", config.DisplayFormat, appConfig.DateFormat.DisplayFormat)
	}
	if config.DBFormat != appConfig.DateFormat.DBFormat {
		t.Errorf("GetDateFormatConfig().DBFormat = %v, want %v", config.DBFormat, appConfig.DateFormat.DBFormat)
	}
}

// TestGetAppConfigFromEnv tests the dependency injection version directly
func TestGetAppConfigFromEnv(t *testing.T) {
	// Test with a mock environment function
	getEnv := func(key string) string {
		switch key {
		case "DATE_INPUT_FORMAT":
			return "mm-dd-yyyy"
		case "DATE_DISPLAY_FORMAT":
			return "yyyy-mm-dd"
		case "DEFAULT_UNIT":
			return "lbs"
		default:
			return ""
		}
	}

	config := GetAppConfigFromEnv(getEnv)

	// Verify the configuration
	expectedConfig := AppConfig{
		DateFormat: DateFormatConfig{
			InputFormat:   "01-02-2006", // mm-dd-yyyy
			DisplayFormat: "2006-01-02", // yyyy-mm-dd
			DBFormat:      DBFormat,
		},
		DefaultUnit: "lbs",
	}

	if config.DateFormat.InputFormat != expectedConfig.DateFormat.InputFormat {
		t.Errorf("InputFormat = %v, want %v", config.DateFormat.InputFormat, expectedConfig.DateFormat.InputFormat)
	}
	if config.DateFormat.DisplayFormat != expectedConfig.DateFormat.DisplayFormat {
		t.Errorf("DisplayFormat = %v, want %v", config.DateFormat.DisplayFormat, expectedConfig.DateFormat.DisplayFormat)
	}
	if config.DefaultUnit != expectedConfig.DefaultUnit {
		t.Errorf("DefaultUnit = %v, want %v", config.DefaultUnit, expectedConfig.DefaultUnit)
	}
}
