package helpers

import (
	"testing"
)

func TestValidateLocale(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantNormalized    string
		wantValid         bool
		description       string
	}{
		// Valid cases - language only
		{
			name:           "valid_en",
			input:          "en",
			wantNormalized: "en",
			wantValid:      true,
			description:    "English language code",
		},
		{
			name:           "valid_zh",
			input:          "zh",
			wantNormalized: "zh",
			wantValid:      true,
			description:    "Chinese language code",
		},
		{
			name:           "valid_ja",
			input:          "ja",
			wantNormalized: "ja",
			wantValid:      true,
			description:    "Japanese language code",
		},

		// Valid cases - language-region
		{
			name:           "valid_zh_TW",
			input:          "zh-TW",
			wantNormalized: "zh-TW",
			wantValid:      true,
			description:    "Chinese (Taiwan)",
		},
		{
			name:           "valid_en_US",
			input:          "en-US",
			wantNormalized: "en-US",
			wantValid:      true,
			description:    "English (United States)",
		},
		{
			name:           "valid_ja_JP",
			input:          "ja-JP",
			wantNormalized: "ja-JP",
			wantValid:      true,
			description:    "Japanese (Japan)",
		},

		// Valid cases - case normalization
		{
			name:           "normalize_lowercase",
			input:          "zh-tw",
			wantNormalized: "zh-TW",
			wantValid:      true,
			description:    "Lowercase should be normalized to proper case",
		},
		{
			name:           "normalize_uppercase",
			input:          "EN-US",
			wantNormalized: "en-US",
			wantValid:      true,
			description:    "Uppercase should be normalized to proper case",
		},

		// Valid cases - language-script-region
		{
			name:           "valid_zh_Hant_TW",
			input:          "zh-Hant-TW",
			wantNormalized: "zh-Hant-TW",
			wantValid:      true,
			description:    "Chinese (Traditional, Taiwan)",
		},
		{
			name:           "valid_zh_Hans_CN",
			input:          "zh-Hans-CN",
			wantNormalized: "zh-Hans-CN",
			wantValid:      true,
			description:    "Chinese (Simplified, China)",
		},

		// Invalid cases
		{
			name:           "invalid_empty",
			input:          "",
			wantNormalized: "",
			wantValid:      false,
			description:    "Empty string should be invalid",
		},
		{
			name:           "invalid_format",
			input:          "invalid-locale-format",
			wantNormalized: "",
			wantValid:      false,
			description:    "Invalid format should be rejected",
		},
		{
			name:           "invalid_language_code",
			input:          "xyz",
			wantNormalized: "",
			wantValid:      false,
			description:    "Non-existent language code should be rejected",
		},
		{
			name:           "invalid_special_chars",
			input:          "en@US",
			wantNormalized: "",
			wantValid:      false,
			description:    "Invalid special characters should be rejected",
		},
		{
			name:           "invalid_multiple_hyphens",
			input:          "en--US",
			wantNormalized: "",
			wantValid:      false,
			description:    "Multiple consecutive hyphens should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNormalized, gotValid := ValidateLocale(tt.input)

			if gotValid != tt.wantValid {
				t.Errorf("ValidateLocale(%q) valid = %v, want %v (%s)",
					tt.input, gotValid, tt.wantValid, tt.description)
			}

			if gotNormalized != tt.wantNormalized {
				t.Errorf("ValidateLocale(%q) normalized = %q, want %q (%s)",
					tt.input, gotNormalized, tt.wantNormalized, tt.description)
			}
		})
	}
}

func TestValidateLocale_MaxLength(t *testing.T) {
	// Test VARCHAR(85) constraint
	longLocale := "en-" + string(make([]byte, 90)) // Create a string longer than 85 chars
	_, valid := ValidateLocale(longLocale)
	if valid {
		t.Error("ValidateLocale should reject locale longer than 85 characters")
	}
}
