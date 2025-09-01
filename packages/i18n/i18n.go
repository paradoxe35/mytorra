package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

//go:embed locales/*.json
var localesFS embed.FS

type Translator struct {
	printers    map[string]*message.Printer
	messages    map[string]map[string]string
	defaultLang string
	catalog     *catalog.Builder
}

// New creates a new Translator with supported languages
func New(defaultLang string, supportedLangs ...string) (*Translator, error) {
	t := &Translator{
		printers:    make(map[string]*message.Printer),
		messages:    make(map[string]map[string]string),
		defaultLang: defaultLang,
		catalog:     catalog.NewBuilder(),
	}

	// Initialize with validation messages
	t.initializeMessages()

	// Load default language
	if err := t.loadLanguage(defaultLang); err != nil {
		return nil, fmt.Errorf("failed to load default language %s: %w", defaultLang, err)
	}

	// Load additional languages
	for _, lang := range supportedLangs {
		if lang != defaultLang {
			if err := t.loadLanguage(lang); err != nil {
				// Log error but don't fail
				fmt.Printf("Warning: failed to load language %s: %v\n", lang, err)
			}
		}
	}

	return t, nil
}

func (t *Translator) loadLanguage(lang string) error {
	// Parse language tag
	tag, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("invalid language tag %s: %w", lang, err)
	}

	// Get messages for this language
	messages, exists := t.messages[lang]
	if !exists {
		// If no messages defined, use default messages
		messages = t.messages["en"]
	}

	// Register messages in catalog
	for key, value := range messages {
		err := t.catalog.Set(tag, key, catalog.String(value))
		if err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to set message %s for %s: %v\n", key, lang, err)
		}
	}

	// Create printer with the catalog
	t.printers[lang] = message.NewPrinter(tag, message.Catalog(t.catalog))

	return nil
}

// T translates a message key for the given language
func (t *Translator) T(lang, key string, args ...interface{}) string {
	// Fallback to default language if not found
	messages, ok := t.messages[lang]
	if !ok {
		lang = t.defaultLang
		messages = t.messages[lang]
	}

	// Get message template
	template, ok := messages[key]
	if !ok {
		// Return key if message not found
		return key
	}

	// Format message with arguments if provided
	if len(args) > 0 {
		return fmt.Sprintf(template, args...)
	}

	return template
}

// TWithContext translates using language from context
func (t *Translator) TWithContext(ctx context.Context, key string, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	if lang == "" {
		lang = t.defaultLang
	}
	return t.T(lang, key, args...)
}

// GetSupportedLanguages returns list of supported languages
func (t *Translator) GetSupportedLanguages() []string {
	langs := make([]string, 0, len(t.messages))
	for lang := range t.messages {
		langs = append(langs, lang)
	}
	return langs
}

// IsSupported checks if a language is supported
func (t *Translator) IsSupported(lang string) bool {
	_, ok := t.messages[lang]
	return ok
}

// Context key for language
type contextKey string

const languageContextKey contextKey = "language"

// WithLanguage adds language to context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, languageContextKey, lang)
}

// GetLanguageFromContext retrieves language from context
func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(languageContextKey).(string); ok {
		return lang
	}
	return ""
}

// ParseAcceptLanguage parses Accept-Language header and returns best match
func ParseAcceptLanguage(header string, supported []string) string {
	if header == "" {
		return ""
	}

	// Simple parsing of Accept-Language header
	// Format: en-US,en;q=0.9,fr;q=0.8
	parts := strings.Split(header, ",")
	for _, part := range parts {
		lang := strings.TrimSpace(strings.Split(part, ";")[0])
		// Convert to lowercase and take first part (en-US -> en)
		lang = strings.ToLower(strings.Split(lang, "-")[0])

		// Check if this language is supported
		for _, sup := range supported {
			if strings.HasPrefix(sup, lang) {
				return sup
			}
		}
	}

	return ""
}

// initializeMessages initializes the message catalog with all translations
func (t *Translator) initializeMessages() {
	// Load messages from embedded JSON files
	t.messages = make(map[string]map[string]string)

	// Try to load from embedded files
	if err := t.loadMessagesFromEmbedded(); err != nil {
		// Fallback to hardcoded messages if embedded files fail
		t.messages = getValidationMessages()
	}
}

// loadMessagesFromEmbedded loads messages from embedded JSON files
func (t *Translator) loadMessagesFromEmbedded() error {
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract language code from filename (e.g., "en.json" -> "en")
		lang := strings.TrimSuffix(entry.Name(), ".json")

		// Read the JSON file
		data, err := localesFS.ReadFile(path.Join("locales", entry.Name()))
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", entry.Name(), err)
			continue
		}

		// Parse JSON
		var messages map[string]string
		if err := json.Unmarshal(data, &messages); err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		t.messages[lang] = messages
	}

	return nil
}

// getValidationMessages returns common validation messages
func getValidationMessages() map[string]map[string]string {
	return map[string]map[string]string{
		"en": {
			"validation.required":           "This field is required",
			"validation.email":              "Please enter a valid email address",
			"validation.phone":              "Please enter a valid phone number",
			"validation.password.min":       "Password must be at least %d characters",
			"validation.password.uppercase": "Password must contain at least one uppercase letter",
			"validation.password.lowercase": "Password must contain at least one lowercase letter",
			"validation.password.number":    "Password must contain at least one number",
			"validation.password.special":   "Password must contain at least one special character",
			"validation.min_length":         "Must be at least %d characters",
			"validation.max_length":         "Must be no more than %d characters",
			"validation.area_invalid":       "Invalid area selected",
		},
		"fr": {
			"validation.required":           "Ce champ est obligatoire",
			"validation.email":              "Veuillez entrer une adresse email valide",
			"validation.phone":              "Veuillez entrer un numéro de téléphone valide",
			"validation.password.min":       "Le mot de passe doit contenir au moins %d caractères",
			"validation.password.uppercase": "Le mot de passe doit contenir au moins une majuscule",
			"validation.password.lowercase": "Le mot de passe doit contenir au moins une minuscule",
			"validation.password.number":    "Le mot de passe doit contenir au moins un chiffre",
			"validation.password.special":   "Le mot de passe doit contenir au moins un caractère spécial",
			"validation.min_length":         "Doit contenir au moins %d caractères",
			"validation.max_length":         "Ne doit pas dépasser %d caractères",
			"validation.area_invalid":       "Zone sélectionnée invalide",
		},
		"sw": {
			"validation.required":           "Sehemu hii ni ya lazima",
			"validation.email":              "Tafadhali weka anwani sahihi ya barua pepe",
			"validation.phone":              "Tafadhali weka nambari sahihi ya simu",
			"validation.password.min":       "Nenosiri lazima liwe na angalau herufi %d",
			"validation.password.uppercase": "Nenosiri lazima liwe na angalau herufi kubwa moja",
			"validation.password.lowercase": "Nenosiri lazima liwe na angalau herufi ndogo moja",
			"validation.password.number":    "Nenosiri lazima liwe na angalau nambari moja",
			"validation.password.special":   "Nenosiri lazima liwe na angalau alama maalum moja",
			"validation.min_length":         "Lazima iwe na angalau herufi %d",
			"validation.max_length":         "Isizidi herufi %d",
			"validation.area_invalid":       "Eneo lililochaguliwa ni batili",
		},
	}
}

// LoadMessagesFromJSON allows loading additional messages from JSON data
func (t *Translator) LoadMessagesFromJSON(lang string, jsonData []byte) error {
	var messages map[string]string
	if err := json.Unmarshal(jsonData, &messages); err != nil {
		return fmt.Errorf("failed to parse JSON messages: %w", err)
	}

	// Merge with existing messages
	if t.messages[lang] == nil {
		t.messages[lang] = make(map[string]string)
	}

	for key, value := range messages {
		t.messages[lang][key] = value
	}

	// Reload language to update catalog
	return t.loadLanguage(lang)
}

// GetPrinter returns the message printer for a specific language
func (t *Translator) GetPrinter(lang string) *message.Printer {
	if printer, ok := t.printers[lang]; ok {
		return printer
	}
	// Fallback to default language
	return t.printers[t.defaultLang]
}
