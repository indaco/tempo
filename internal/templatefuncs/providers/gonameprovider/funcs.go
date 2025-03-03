package gonameprovider

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
)

// ToGoPackageName converts a string into a valid Go package name.
// It handles kebab-case, snake_case, camelCase, and PascalCase.
func ToGoPackageName(input string) string {
	// Replace hyphens and spaces with underscores
	input = strings.ReplaceAll(input, "-", "_")
	input = strings.ReplaceAll(input, " ", "_")

	// Convert PascalCase or camelCase to snake_case using regex
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	input = re.ReplaceAllString(input, `${1}_${2}`)

	// Replace invalid characters (non-alphanumeric, non-underscore, and non-slash) with underscores
	input = regexp.MustCompile(`[^a-zA-Z0-9_/]`).ReplaceAllString(input, "_")

	// Convert the string to lowercase
	input = strings.ToLower(input)

	// Replace multiple underscores with a single underscore (excluding slashes)
	input = regexp.MustCompile(`_+`).ReplaceAllString(input, "_")

	// Remove leading and trailing underscores
	input = strings.Trim(input, "_")

	// If the first character is a digit, prepend an underscore
	if len(input) > 0 && input[0] >= '0' && input[0] <= '9' {
		input = "_" + input
	}

	return input
}

// ToGoExportedName converts a string to a valid exported Go function name (PascalCase).
func ToGoExportedName(input string) string {
	// Remove invalid characters, keeping only letters, digits, and hyphens
	input = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(input, " ")

	// Split into words using spaces or hyphens
	words := regexp.MustCompile(`[ -]+`).Split(input, -1)

	// Remove empty words caused by extra delimiters
	nonEmptyWords := make([]string, 0, len(words))
	for _, word := range words {
		if !textprovider.IsEmptyString(word) {
			nonEmptyWords = append(nonEmptyWords, word)
		}
	}

	// Capitalize each word
	for i := range nonEmptyWords {
		nonEmptyWords[i] = textprovider.TitleCase(nonEmptyWords[i])
	}

	// Join words together in PascalCase
	result := strings.Join(nonEmptyWords, "")

	// Ensure the name starts with an uppercase letter and doesn't start with a digit
	if len(result) > 0 && unicode.IsDigit(rune(result[0])) {
		result = "Invalid" + result
	}

	return result
}

// ToGoUnexportedName converts a string to a valid unexported Go function name (camelCase).
func ToGoUnexportedName(input string) string {
	// Normalize input: replace non-alphanumeric characters with spaces
	input = regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(input, " ")

	// Split into words using camelCase splitting and space normalization
	words := regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(input, "${1} ${2}")
	wordList := regexp.MustCompile(`\s+`).Split(words, -1)

	// Filter out empty words caused by extra delimiters
	nonEmptyWords := make([]string, 0, len(wordList))
	for _, word := range wordList {
		if !textprovider.IsEmptyString(word) {
			nonEmptyWords = append(nonEmptyWords, word)
		}
	}

	if len(nonEmptyWords) == 0 {
		return ""
	}

	// Construct camelCase: first word lowercase, subsequent words capitalized
	result := strings.ToLower(nonEmptyWords[0])
	for _, word := range nonEmptyWords[1:] {
		result += textprovider.TitleCase(word)
	}

	// Handle leading digits
	if unicode.IsDigit(rune(result[0])) {
		result = "invalid" + textprovider.TitleCase(result)
	}

	return result
}
