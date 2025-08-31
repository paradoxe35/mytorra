package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// I18nValidator validates that all i18n keys used in code exist in locale files
type I18nValidator struct {
	localeKeys  map[string]map[string]bool // lang -> key -> exists
	usedKeys    map[string][]string        // key -> []file_locations
	missingKeys []string
	unusedKeys  []string
	localesDir  string
	projectRoot string
}

// NewI18nValidator creates a new validator instance
func NewI18nValidator(projectRoot, localesDir string) *I18nValidator {
	return &I18nValidator{
		localeKeys:  make(map[string]map[string]bool),
		usedKeys:    make(map[string][]string),
		localesDir:  localesDir,
		projectRoot: projectRoot,
	}
}

// LoadLocaleFiles loads all locale JSON files
func (v *I18nValidator) LoadLocaleFiles() error {
	entries, err := os.ReadDir(v.localesDir)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")
		filePath := filepath.Join(v.localesDir, entry.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		var messages map[string]interface{}
		if err := json.Unmarshal(data, &messages); err != nil {
			return fmt.Errorf("failed to parse %s: %w", filePath, err)
		}

		v.localeKeys[lang] = make(map[string]bool)
		v.flattenKeys("", messages, v.localeKeys[lang])

		fmt.Printf("Loaded %d keys from %s\n", len(v.localeKeys[lang]), entry.Name())
	}

	return nil
}

// flattenKeys recursively flattens nested JSON keys
func (v *I18nValidator) flattenKeys(prefix string, obj map[string]interface{}, result map[string]bool) {
	for key, value := range obj {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch val := value.(type) {
		case map[string]interface{}:
			v.flattenKeys(fullKey, val, result)
		case string:
			result[fullKey] = true
		default:
			// Handle other types as strings
			result[fullKey] = true
		}
	}
}

// ScanCodeForI18nKeys scans Go files for i18n key usage
func (v *I18nValidator) ScanCodeForI18nKeys() error {
	return filepath.WalkDir(v.projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, node_modules, .git, etc.
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" ||
				name == "dist" || name == "build" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return v.scanGoFile(path)
	})
}

// scanGoFile scans a single Go file for i18n key usage
func (v *I18nValidator) scanGoFile(filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	// Look for i18n function calls
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			v.extractI18nKeys(x, filePath, fset)
		}
		return true
	})

	return nil
}

// extractI18nKeys extracts i18n keys from function calls
func (v *I18nValidator) extractI18nKeys(call *ast.CallExpr, filePath string, fset *token.FileSet) {
	// Look for calls like: translator.T("key"), translator.TWithContext(ctx, "key"), i18n.T("key"), T("key")
	var funcName string
	var receiverName string

	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		funcName = fun.Sel.Name
		// Check if receiver is "translator" or similar
		if ident, ok := fun.X.(*ast.Ident); ok {
			receiverName = ident.Name
		}
	case *ast.Ident:
		funcName = fun.Name
	default:
		return
	}

	// Check if this is an i18n function
	if !v.isI18nFunction(funcName) {
		return
	}

	// Skip if receiver doesn't look like a translator
	if receiverName != "" && !v.isTranslatorReceiver(receiverName) {
		return
	}

	// Extract string literals from arguments
	for i, arg := range call.Args {
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			// Remove quotes from string literal
			key := strings.Trim(lit.Value, `"`)
			key = strings.Trim(key, `'`)

			// Skip empty keys
			if key == "" {
				continue
			}

			// Skip if it's likely a language code (first argument for some functions)
			if i == 0 && v.isLanguageCode(key) && funcName == "T" {
				continue
			}

			// For TWithContext, the key is the second argument (after context)
			if funcName == "TWithContext" && i != 1 {
				continue
			}

			// For T function when called on translator object, the first string is usually a language code
			// and the second is the key. But if first arg is not a language code, it's the key.
			if funcName == "T" {
				if i == 0 && !v.isLanguageCode(key) {
					// This is the key (first argument is not a language code)
				} else if i == 1 {
					// This is the key (second argument after language)
				} else {
					continue
				}
			}

			// Skip format specifiers (like %d, %s)
			if strings.Contains(key, "%") && len(key) <= 3 {
				continue
			}

			// Skip if key looks like a variable name or placeholder
			if !strings.Contains(key, ".") && !strings.Contains(key, "_") {
				// Most i18n keys have dots or underscores
				if len(key) < 5 { // Skip very short keys that might be placeholders
					continue
				}
			}

			pos := fset.Position(lit.Pos())
			location := fmt.Sprintf("%s:%d:%d", filePath, pos.Line, pos.Column)
			v.usedKeys[key] = append(v.usedKeys[key], location)
		}
	}
}

// isI18nFunction checks if a function name is an i18n function
func (v *I18nValidator) isI18nFunction(name string) bool {
	i18nFunctions := []string{"T", "TWithContext", "Translate", "GetMessage", "Localize", "MustLocalize"}
	for _, fn := range i18nFunctions {
		if name == fn {
			return true
		}
	}
	return false
}

// isTranslatorReceiver checks if a receiver name looks like a translator
func (v *I18nValidator) isTranslatorReceiver(name string) bool {
	translatorNames := []string{"translator", "trans", "t", "i18n", "l10n", "locale", "localizer"}
	lowerName := strings.ToLower(name)
	for _, tn := range translatorNames {
		if strings.Contains(lowerName, tn) {
			return true
		}
	}
	// Also check for common service/struct receiver names that might have translators
	if strings.HasSuffix(lowerName, "service") || strings.HasSuffix(lowerName, "handler") ||
		strings.HasSuffix(lowerName, "controller") || name == "s" || name == "h" || name == "c" {
		return true
	}
	return false
}

// isLanguageCode checks if a string looks like a language code
func (v *I18nValidator) isLanguageCode(s string) bool {
	// Common language codes
	langCodes := []string{"en", "fr", "es", "de", "it", "pt", "sw", "ar", "zh", "ja", "ko", "ru", "hi"}
	for _, code := range langCodes {
		if s == code {
			return true
		}
	}
	// Check for locale format (e.g., en-US, fr-CA)
	matched, _ := regexp.MatchString(`^[a-z]{2,3}(-[A-Z]{2})?$`, s)
	return matched
}

// ValidateKeys validates that all used keys exist in locale files
func (v *I18nValidator) ValidateKeys() {
	// Find missing keys
	for key := range v.usedKeys {
		found := false
		for lang := range v.localeKeys {
			if v.localeKeys[lang][key] {
				found = true
				break
			}
		}
		if !found {
			v.missingKeys = append(v.missingKeys, key)
		}
	}

	// Find unused keys (keys in locale files but not used in code)
	for lang, keys := range v.localeKeys {
		for key := range keys {
			if _, used := v.usedKeys[key]; !used {
				v.unusedKeys = append(v.unusedKeys, fmt.Sprintf("%s:%s", lang, key))
			}
		}
	}

	sort.Strings(v.missingKeys)
	sort.Strings(v.unusedKeys)
}

// PrintReport prints the validation report
func (v *I18nValidator) PrintReport() {
	fmt.Printf("\n=== I18n Validation Report ===\n")
	fmt.Printf("Scanned %d unique keys used in code\n", len(v.usedKeys))

	totalLocaleKeys := 0
	for _, keys := range v.localeKeys {
		totalLocaleKeys += len(keys)
	}
	fmt.Printf("Found %d total keys in locale files\n", totalLocaleKeys)

	if len(v.missingKeys) > 0 {
		fmt.Printf("\n❌ Missing Keys (%d):\n", len(v.missingKeys))
		for _, key := range v.missingKeys {
			fmt.Printf("  - %s\n", key)
			if locations, ok := v.usedKeys[key]; ok {
				for _, loc := range locations {
					fmt.Printf("    Used in: %s\n", loc)
				}
			}
		}
	}

	if len(v.unusedKeys) > 0 {
		fmt.Printf("\n⚠️  Unused Keys (%d):\n", len(v.unusedKeys))
		for _, key := range v.unusedKeys {
			fmt.Printf("  - %s\n", key)
		}
	}

	if len(v.missingKeys) == 0 && len(v.unusedKeys) == 0 {
		fmt.Printf("\n✅ All i18n keys are valid!\n")
	}
}

// HasErrors returns true if there are validation errors
func (v *I18nValidator) HasErrors() bool {
	return len(v.missingKeys) > 0
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: i18n-validator <project-root> [locales-dir]")
		fmt.Println("Example: i18n-validator . packages/i18n/locales")
		os.Exit(1)
	}

	projectRoot := os.Args[1]
	localesDir := "packages/i18n/locales"
	if len(os.Args) > 2 {
		localesDir = os.Args[2]
	}

	validator := NewI18nValidator(projectRoot, localesDir)

	fmt.Printf("Loading locale files from %s...\n", localesDir)
	if err := validator.LoadLocaleFiles(); err != nil {
		fmt.Printf("Error loading locale files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Scanning Go files in %s...\n", projectRoot)
	if err := validator.ScanCodeForI18nKeys(); err != nil {
		fmt.Printf("Error scanning code: %v\n", err)
		os.Exit(1)
	}

	validator.ValidateKeys()
	validator.PrintReport()

	if validator.HasErrors() {
		os.Exit(1)
	}
}
