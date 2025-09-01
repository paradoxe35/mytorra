# i18n Key Validator

A Go tool that validates internationalization (i18n) key usage across the MyTorra platform codebase.

## Features

- **Missing Key Detection**: Finds i18n keys used in code but missing from locale files
- **Unused Key Detection**: Identifies keys defined in locale files but not used in code
- **Multi-language Support**: Validates across all locale files (en.json, fr.json, etc.)
- **Detailed Reporting**: Shows exact file locations where keys are used
- **CI/CD Integration**: Designed to run in GitHub Actions workflows

## Usage

### Command Line

```bash
# Build the tool
make i18n-build

# Validate i18n keys (fails on missing keys)
make i18n-validate

# Check for unused keys (non-failing)
make i18n-unused

# Direct usage
./tools/i18n-validator/i18n-validator <project-root> [locales-dir]
```

### Examples

```bash
# Validate from project root
./tools/i18n-validator/i18n-validator . packages/i18n/locales

# Validate specific directory
./tools/i18n-validator/i18n-validator services/core packages/i18n/locales
```

## Supported i18n Function Patterns

The validator recognizes these function call patterns:

```go
// Direct translation calls
translator.T("validation.required")
translator.T("en", "validation.required")
translator.TWithContext(ctx, "user.welcome")

// Package-level functions
i18n.T("error.not_found")
T("success.saved")
```

## Locale File Format

Locale files should be JSON with nested keys:

```json
{
  "validation": {
    "required": "This field is required",
    "email": "Please enter a valid email"
  },
  "user": {
    "welcome": "Welcome, %s!",
    "profile": {
      "updated": "Profile updated successfully"
    }
  }
}
```

Keys are flattened using dot notation: `validation.required`, `user.profile.updated`

## Output Example

```
=== I18n Validation Report ===
Scanned 45 unique keys used in code
Found 52 total keys in locale files

❌ Missing Keys (2):
  - validation.password.complexity
    Used in: services/core/internal/adapters/http/handlers/auth_handler.go:123:45
  - user.settings.privacy
    Used in: services/core/internal/application/services/user_service.go:89:12

⚠️  Unused Keys (3):
  - en:validation.old_field
  - fr:user.deprecated_message
  - en:error.legacy_code
```

## Integration

### GitHub Actions

The tool is integrated into:
- **CI Pipeline** (`ci.yml`): Validates i18n keys on every push
- **PR Validation** (`pr-validation.yml`): Blocks PRs with missing keys

### Make Commands

- `make validate`: Includes i18n validation in full validation suite
- `make setup`: Builds the tool during project setup

## Error Codes

- **Exit 0**: All validations passed
- **Exit 1**: Missing keys found (validation failure)

Note: Unused keys generate warnings but don't cause validation failure.

## Development

### Building

```bash
cd tools/i18n-validator
go build -o i18n-validator .
```

### Testing

```bash
cd tools/i18n-validator
go test ./...
```

## Configuration

The tool uses these defaults:
- **Project Root**: Current directory (`.`)
- **Locales Directory**: `packages/i18n/locales`
- **Supported Extensions**: `.go` files (excluding `_test.go`)
- **Excluded Directories**: `vendor`, `node_modules`, `.git`, `dist`, `build`

## Troubleshooting

### Common Issues

1. **"No locale files found"**
   - Ensure locale files exist in `packages/i18n/locales/`
   - Check file names match pattern `*.json`

2. **"False positives for missing keys"**
   - Verify key names match exactly (case-sensitive)
   - Check for typos in function calls

3. **"Keys not detected in code"**
   - Ensure using supported function patterns
   - Check if files are in excluded directories

### Debug Mode

For detailed scanning information, modify the tool to enable debug logging:

```go
// Add to main() function
fmt.Printf("Scanning file: %s\n", filePath)
```
