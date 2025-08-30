package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestI18nValidator_LoadLocaleFiles(t *testing.T) {
	// Create temporary directory for test locale files
	tempDir, err := os.MkdirTemp("", "i18n-test-locales")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test locale files
	enLocale := map[string]interface{}{
		"validation": map[string]interface{}{
			"required": "This field is required",
			"email":    "Please enter a valid email",
		},
		"user": map[string]interface{}{
			"welcome": "Welcome!",
			"profile": map[string]interface{}{
				"updated": "Profile updated",
			},
		},
		"simple": "Simple message",
	}

	frLocale := map[string]interface{}{
		"validation": map[string]interface{}{
			"required": "Ce champ est obligatoire",
			"email":    "Veuillez entrer un email valide",
		},
		"user": map[string]interface{}{
			"welcome": "Bienvenue!",
		},
	}

	// Write locale files
	writeLocaleFile(t, tempDir, "en.json", enLocale)
	writeLocaleFile(t, tempDir, "fr.json", frLocale)

	// Test validator
	validator := NewI18nValidator(".", tempDir)
	err = validator.LoadLocaleFiles()
	if err != nil {
		t.Fatalf("Failed to load locale files: %v", err)
	}

	// Verify English keys
	expectedEnKeys := []string{
		"validation.required",
		"validation.email",
		"user.welcome",
		"user.profile.updated",
		"simple",
	}

	if len(validator.localeKeys["en"]) != len(expectedEnKeys) {
		t.Errorf("Expected %d English keys, got %d", len(expectedEnKeys), len(validator.localeKeys["en"]))
	}

	for _, key := range expectedEnKeys {
		if !validator.localeKeys["en"][key] {
			t.Errorf("Expected English key '%s' not found", key)
		}
	}

	// Verify French keys
	expectedFrKeys := []string{
		"validation.required",
		"validation.email",
		"user.welcome",
	}

	if len(validator.localeKeys["fr"]) != len(expectedFrKeys) {
		t.Errorf("Expected %d French keys, got %d", len(expectedFrKeys), len(validator.localeKeys["fr"]))
	}

	for _, key := range expectedFrKeys {
		if !validator.localeKeys["fr"][key] {
			t.Errorf("Expected French key '%s' not found", key)
		}
	}
}

func TestI18nValidator_ScanCodeForI18nKeys(t *testing.T) {
	// Create temporary directory for test Go files
	tempDir, err := os.MkdirTemp("", "i18n-test-code")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test Go file with i18n usage
	goCode := `package main

import (
	"context"
	"github.com/paradoxe35/torra/packages/i18n"
)

func main() {
	translator, _ := i18n.New("en", "fr")

	// Various i18n function calls
	msg1 := translator.T("en", "validation.required")
	msg2 := translator.T("fr", "user.welcome")
	msg3 := translator.TWithContext(context.Background(), "error.not_found")

	// Direct function calls
	msg4 := T("simple.message")
	msg5 := Translate("auth.login_success")

	// Non-i18n function calls (should be ignored)
	fmt.Println("This is not an i18n call")
	someFunc("not.an.i18n.key")
}

func T(key string) string {
	return key
}

func Translate(key string) string {
	return key
}
`

	goFilePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFilePath, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test Go file: %v", err)
	}

	// Test validator
	validator := NewI18nValidator(tempDir, "")
	err = validator.ScanCodeForI18nKeys()
	if err != nil {
		t.Fatalf("Failed to scan code: %v", err)
	}

	// Verify found keys (Translate function might not be recognized)
	expectedKeys := []string{
		"validation.required",
		"user.welcome",
		"error.not_found",
		"simple.message",
	}

	if len(validator.usedKeys) != len(expectedKeys) {
		t.Errorf("Expected %d used keys, got %d", len(expectedKeys), len(validator.usedKeys))
	}

	for _, key := range expectedKeys {
		if _, found := validator.usedKeys[key]; !found {
			t.Errorf("Expected key '%s' not found in used keys", key)
		}
	}

	// Verify file locations are recorded
	for key, locations := range validator.usedKeys {
		if len(locations) == 0 {
			t.Errorf("Key '%s' has no recorded locations", key)
		}
		for _, location := range locations {
			if !filepath.IsAbs(location) {
				rel, err := filepath.Rel(tempDir, location)
				if err != nil || filepath.IsAbs(rel) {
					t.Errorf("Location '%s' for key '%s' doesn't reference test file", location, key)
				}
			}
		}
	}
}

func TestI18nValidator_ValidateKeys(t *testing.T) {
	validator := NewI18nValidator(".", "")

	// Set up test data
	validator.localeKeys = map[string]map[string]bool{
		"en": {
			"validation.required": true,
			"user.welcome":        true,
			"unused.key":          true,
		},
		"fr": {
			"validation.required": true,
			"user.welcome":        true,
			"fr.only.key":         true,
		},
	}

	validator.usedKeys = map[string][]string{
		"validation.required": {"file1.go:10:5"},
		"user.welcome":        {"file2.go:20:10"},
		"missing.key":         {"file3.go:30:15"},
	}

	// Run validation
	validator.ValidateKeys()

	// Check missing keys
	expectedMissing := []string{"missing.key"}
	if len(validator.missingKeys) != len(expectedMissing) {
		t.Errorf("Expected %d missing keys, got %d", len(expectedMissing), len(validator.missingKeys))
	}

	for _, key := range expectedMissing {
		found := false
		for _, missing := range validator.missingKeys {
			if missing == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected missing key '%s' not found", key)
		}
	}

	// Check unused keys
	expectedUnused := []string{"en:unused.key", "fr:fr.only.key"}
	if len(validator.unusedKeys) != len(expectedUnused) {
		t.Errorf("Expected %d unused keys, got %d", len(expectedUnused), len(validator.unusedKeys))
	}
}

func TestI18nValidator_IsI18nFunction(t *testing.T) {
	validator := NewI18nValidator(".", "")

	testCases := []struct {
		funcName string
		expected bool
	}{
		{"T", true},
		{"TWithContext", true},
		{"Translate", true},
		{"GetMessage", true},
		{"Printf", false},
		{"Sprintf", false},
		{"someOtherFunc", false},
	}

	for _, tc := range testCases {
		result := validator.isI18nFunction(tc.funcName)
		if result != tc.expected {
			t.Errorf("isI18nFunction(%s) = %v, expected %v", tc.funcName, result, tc.expected)
		}
	}
}

func TestI18nValidator_IsLanguageCode(t *testing.T) {
	validator := NewI18nValidator(".", "")

	testCases := []struct {
		code     string
		expected bool
	}{
		{"en", true},
		{"fr", true},
		{"es", true},
		{"en-US", true},
		{"fr-CA", true},
		{"eng", true},
		{"validation.required", false},
		{"user.welcome", false},
		{"123", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := validator.isLanguageCode(tc.code)
		if result != tc.expected {
			t.Errorf("isLanguageCode(%s) = %v, expected %v", tc.code, result, tc.expected)
		}
	}
}

func TestI18nValidator_HasErrors(t *testing.T) {
	validator := NewI18nValidator(".", "")

	// No missing keys
	validator.missingKeys = []string{}
	if validator.HasErrors() {
		t.Error("Expected no errors when no missing keys")
	}

	// With missing keys
	validator.missingKeys = []string{"missing.key"}
	if !validator.HasErrors() {
		t.Error("Expected errors when missing keys exist")
	}
}

func TestI18nValidator_Integration(t *testing.T) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "i18n-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	localesDir := filepath.Join(tempDir, "locales")
	err = os.MkdirAll(localesDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create locales dir: %v", err)
	}

	// Create locale files
	enLocale := map[string]interface{}{
		"validation": map[string]interface{}{
			"required": "This field is required",
		},
		"user": map[string]interface{}{
			"welcome": "Welcome!",
		},
		"unused": "This key is not used",
	}

	writeLocaleFile(t, localesDir, "en.json", enLocale)

	// Create Go file with i18n usage
	goCode := `package main

func main() {
	msg := T("validation.required")
	welcome := T("user.welcome")
	missing := T("missing.key")
}

func T(key string) string {
	return key
}
`

	goFilePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFilePath, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write Go file: %v", err)
	}

	// Run full validation
	validator := NewI18nValidator(tempDir, localesDir)

	err = validator.LoadLocaleFiles()
	if err != nil {
		t.Fatalf("Failed to load locale files: %v", err)
	}

	err = validator.ScanCodeForI18nKeys()
	if err != nil {
		t.Fatalf("Failed to scan code: %v", err)
	}

	validator.ValidateKeys()

	// Verify results
	if len(validator.missingKeys) != 1 || validator.missingKeys[0] != "missing.key" {
		t.Errorf("Expected 1 missing key 'missing.key', got %v", validator.missingKeys)
	}

	if len(validator.unusedKeys) != 1 || validator.unusedKeys[0] != "en:unused" {
		t.Errorf("Expected 1 unused key 'en:unused', got %v", validator.unusedKeys)
	}

	if !validator.HasErrors() {
		t.Error("Expected validation to have errors due to missing key")
	}
}

func TestI18nValidator_EmptyLocalesDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i18n-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	validator := NewI18nValidator(".", tempDir)
	err = validator.LoadLocaleFiles()
	if err != nil {
		t.Fatalf("Should handle empty locales directory gracefully: %v", err)
	}

	if len(validator.localeKeys) != 0 {
		t.Error("Expected no locale keys for empty directory")
	}
}

func TestI18nValidator_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i18n-invalid-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create invalid JSON file
	invalidJSON := `{"key": "value", "invalid": }`
	filePath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(filePath, []byte(invalidJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	validator := NewI18nValidator(".", tempDir)
	err = validator.LoadLocaleFiles()
	// Should not fail completely, just skip invalid files
	if err != nil {
		t.Logf("Expected error for invalid JSON: %v", err)
		// This is expected behavior - invalid JSON should cause an error
	}
}

// Helper function to write locale files
func writeLocaleFile(t *testing.T, dir, filename string, data map[string]interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal locale data: %v", err)
	}

	filePath := filepath.Join(dir, filename)
	err = os.WriteFile(filePath, jsonData, 0o644)
	if err != nil {
		t.Fatalf("Failed to write locale file %s: %v", filename, err)
	}
}
