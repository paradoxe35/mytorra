package i18n_test

import (
	"context"
	"testing"

	"github.com/paradoxe35/mytorra/packages/i18n"
)

func TestTranslator_New(t *testing.T) {
	translator, err := i18n.New("en", "fr", "sw")
	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	if translator == nil {
		t.Fatal("Translator should not be nil")
	}

	// Test supported languages
	supported := translator.GetSupportedLanguages()
	if len(supported) == 0 {
		t.Error("Should have at least one supported language")
	}

	// Test if English is supported
	if !translator.IsSupported("en") {
		t.Error("English should be supported")
	}
}

func TestTranslator_T(t *testing.T) {
	translator, err := i18n.New("en", "fr")
	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	tests := []struct {
		name     string
		lang     string
		key      string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple translation",
			lang:     "en",
			key:      "validation.required",
			expected: "This field is required",
		},
		{
			name:     "translation with args",
			lang:     "en",
			key:      "validation.min_length",
			args:     []interface{}{10},
			expected: "Must be at least 10 characters",
		},
		{
			name:     "french translation",
			lang:     "fr",
			key:      "validation.required",
			expected: "Ce champ est obligatoire",
		},
		{
			name:     "missing key returns key",
			lang:     "en",
			key:      "non.existent.key",
			expected: "non.existent.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.T(tt.lang, tt.key, tt.args...)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTranslator_TWithContext(t *testing.T) {
	translator, err := i18n.New("en", "fr")
	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	// Test with language in context
	ctx := i18n.WithLanguage(context.Background(), "fr")
	result := translator.TWithContext(ctx, "validation.required")
	expected := "Ce champ est obligatoire"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test without language in context (should use default)
	ctx2 := context.Background()
	result2 := translator.TWithContext(ctx2, "validation.required")
	expected2 := "This field is required"

	if result2 != expected2 {
		t.Errorf("Expected '%s', got '%s'", expected2, result2)
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	supported := []string{"en", "fr", "sw"}

	tests := []struct {
		header   string
		expected string
	}{
		{
			header:   "en-US,en;q=0.9,fr;q=0.8",
			expected: "en",
		},
		{
			header:   "fr-FR,fr;q=0.9",
			expected: "fr",
		},
		{
			header:   "de-DE,de;q=0.9,en;q=0.8",
			expected: "en", // German not supported, fallback to English
		},
		{
			header:   "",
			expected: "",
		},
		{
			header:   "zh-CN",
			expected: "", // Chinese not supported
		},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			result := i18n.ParseAcceptLanguage(tt.header, supported)
			if result != tt.expected {
				t.Errorf("For header '%s', expected '%s', got '%s'", tt.header, tt.expected, result)
			}
		})
	}
}

func TestLoadMessagesFromJSON(t *testing.T) {
	translator, err := i18n.New("en")
	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	// Load additional messages
	customMessages := `{
		"custom.message": "This is a custom message",
		"custom.greeting": "Hello, %s!"
	}`

	err = translator.LoadMessagesFromJSON("en", []byte(customMessages))
	if err != nil {
		t.Fatalf("Failed to load custom messages: %v", err)
	}

	// Test custom message
	result := translator.T("en", "custom.message")
	expected := "This is a custom message"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test custom message with args
	result2 := translator.T("en", "custom.greeting", "World")
	expected2 := "Hello, World!"
	if result2 != expected2 {
		t.Errorf("Expected '%s', got '%s'", expected2, result2)
	}
}

func BenchmarkTranslator_T(b *testing.B) {
	translator, _ := i18n.New("en", "fr")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = translator.T("en", "validation.required")
	}
}

func BenchmarkTranslator_TWithArgs(b *testing.B) {
	translator, _ := i18n.New("en", "fr")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = translator.T("en", "validation.min_length", 10)
	}
}

func BenchmarkTranslator_TWithContext(b *testing.B) {
	translator, _ := i18n.New("en", "fr")
	ctx := i18n.WithLanguage(context.Background(), "fr")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = translator.TWithContext(ctx, "validation.required")
	}
}
